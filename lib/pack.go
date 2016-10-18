package icepacker

import (
	"os"
	"path/filepath"
	"regexp"
)

// PackSettings records the settings of the packing
type PackSettings struct {
	SourceDir      string
	TargetFilename string
	Includes       string
	Excludes       string
	Compression    byte
	Encryption     byte
	Cipher         CipherSettings
	OnProgress     chan ProgressState
	OnFinish       chan FinishResult
}

// Progress push a success ProgressState instance to the OnProgress channel.
func (this *PackSettings) Progress(total, index int, filename string) {
	if this.OnProgress != nil {
		this.OnProgress <- ProgressState{nil, total, index, filename}
	}
}

// ProgressError push an error ProgressState instance to the OnProgress channel.
func (this *PackSettings) ProgressError(err error, filename string) {
	if this.OnProgress != nil {
		this.OnProgress <- ProgressState{err, 0, 0, filename}
	}
}

// Finish returns a success FinishResult instance and put to the OnFinish channel if it's not nil
func (this *PackSettings) Finish(err error, fileCount int64, size int64, dupCount int, dupSize int64) FinishResult {
	ret := FinishResult{err, fileCount, size, dupCount, dupSize}
	if this.OnFinish != nil {
		this.OnFinish <- ret
	}
	return ret
}

// Finish returns an errored FinishResult instance and put to the OnFinish channel if it's not nil
func (this *PackSettings) FinishError(err error) FinishResult {
	ret := FinishResult{Err: err}
	if this.OnFinish != nil {
		this.OnFinish <- ret
	}
	return ret
}

// collectFiles walk directories recursively and collect files considering include and exclude filters
func collectFiles(settings PackSettings) []string {
	files := []string{}

	filepath.Walk(FixPath(settings.SourceDir), func(path string, f os.FileInfo, err error) error {
		if err != nil {
			if f != nil {
				settings.ProgressError(err, f.Name())
			} else {
				settings.ProgressError(err, path)
			}
			return err
		}

		if f != nil {
			needAppend := true

			// Matching include filter
			if settings.Includes != "" && !f.IsDir() {
				if matched, _ := regexp.MatchString(settings.Includes, path); matched {
				} else {
					needAppend = false
				}
			}

			// Unmatching exclude filter
			if needAppend && settings.Excludes != "" {
				if matched, _ := regexp.MatchString(settings.Excludes, path); !matched {
				} else {
					needAppend = false
				}
			}

			if needAppend {
				if !f.IsDir() {
					files = append(files, path)
				}
			} else {
				if f.IsDir() {
					return filepath.SkipDir
				}
			}

		}
		return nil
	})
	return files
}

// Pack bundles the files of the source directory to the target package file.
func Pack(settings PackSettings) FinishResult {

	// Hash the cipher key
	shaKey := HashingKey(settings.Cipher)

	// Create a new bundle
	bundle, err := CreateBundle(settings.TargetFilename, BundleSettings{Compression: settings.Compression, Encryption: settings.Encryption, CipherKey: shaKey})
	if err != nil {
		return settings.FinishError(err)
	}
	defer bundle.Close()

	// Get info from source
	sourceInfo, err := os.Stat(settings.SourceDir)
	if err != nil {
		return settings.FinishError(err)
	}

	files := []string{}
	if sourceInfo.IsDir() {
		// Collect files from source directory
		files = collectFiles(settings)
	} else {
		// SourceDir is a file, not a directory
		files = append(files, settings.SourceDir)
	}

	fileCount := len(files)

	for i, file := range files {

		func(i int, file string) {

			var relativePath string
			if file == settings.SourceDir {
				// SourceDir is a file
				relativePath = filepath.Base(FixPath(settings.SourceDir))
			} else {
				relativePath, _ = filepath.Rel(FixPath(settings.SourceDir), file)
			}

			// Update progress state at every 100th file
			if i%100 == 0 {
				settings.Progress(fileCount, i, relativePath)
			}

			_, err := bundle.AddFile(relativePath, file)
			if err != nil {
				settings.ProgressError(err, relativePath)
				return
			}

		}(i, file)
	}

	// Update progress state
	settings.Progress(fileCount, fileCount, "")

	err = bundle.Finalize()
	if err != nil {
		return settings.FinishError(err)
	}

	// Process finished
	return settings.Finish(nil, bundle.FAT.Count, bundle.Footer.PackSize, bundle.DupCount, bundle.DupSize)
}
