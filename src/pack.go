package icepacker

import (
	"crypto/sha512"
	"io/ioutil"
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

// findDuplication finds the duplicated file contents by hash of content.
func findDuplicate(newItem *FATItem, fat *FAT) *FATItem {
	for _, item := range fat.Items {
		if newItem.Hash == item.Hash && newItem.OrigSize == item.OrigSize {
			return &item
		}
	}
	return nil
}

// Pack bundles the files of the source directory to the target package file.
func Pack(settings PackSettings) FinishResult {

	// Create folders for target file
	err := os.MkdirAll(filepath.Dir(settings.TargetFilename), DEFAULT_PERMISSION)
	if err != nil {
		return settings.FinishError(err)
	}

	// Create target package file
	dest, err := os.Create(settings.TargetFilename)
	if err != nil {
		return settings.FinishError(err)
	}
	defer dest.Close()

	// Create a new header for package file
	header := NewHeader(settings.Encryption, settings.Compression)
	err = header.Write(dest)
	if err != nil {
		return settings.FinishError(err)
	}

	// Hash the cipher key
	shaKey := HashingKey(settings.Cipher)

	// Get info from source
	sourceInfo, err := os.Stat(settings.SourceDir)
	if err != nil {
		return settings.FinishError(err)
	}

	files := []string{}
	if sourceInfo.IsDir() {
		// Walk source directory & collect files
		filepath.Walk(FixPath(settings.SourceDir), func(path string, f os.FileInfo, err error) error {
			if err != nil && f != nil {
				settings.ProgressError(err, f.Name())
				return err
			}

			if f != nil && !f.IsDir() {
				needAppend := true

				// Matching include filter
				if settings.Includes != "" {
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
					files = append(files, path)
				}

			} else {
				// TODO: match directories too
				/*dir := filepath.Base(path)
				  for _, d := range ignoreDirs {
				      if d == dir {
				          return filepath.SkipDir
				      }
				  }*/
			}
			return nil
		})
	} else {
		// SourceDir is a file, not a directory
		files = append(files, settings.SourceDir)
	}

	// Create FAT
	fat := FAT{Count: 0, Size: 0}
	fileCount := len(files)

	// Duplicate counters
	dupCount := 0
	dupSize := int64(0)

	for i, file := range files {

		func(i int, file string) {

			var relativePath string

			if file == settings.SourceDir {
				// SourceDir is a file
				relativePath = filepath.Base(FixPath(settings.SourceDir))
			} else {
				relativePath, _ = filepath.Rel(FixPath(settings.SourceDir), file)
			}

			// Open source file
			f, err := os.Open(file)
			if err != nil {
				settings.ProgressError(err, relativePath)
				return
			}
			defer f.Close()

			// Get file info
			fileInfo, err := f.Stat()
			if err != nil {
				settings.ProgressError(err, relativePath)
				return
			}

			// Update progress state at every 100. file
			if i%100 == 0 {
				settings.Progress(fileCount, i, relativePath)
			}

			fileMode := fileInfo.Mode()
			fileSize := fileInfo.Size()

			// Create a FAT item
			item := FATItem{Path: filepath.ToSlash(relativePath), Offset: fat.Size, OrigSize: fileSize, MTime: fileInfo.ModTime().UnixNano(), Mode: uint32(fileMode), Perm: uint32(fileMode.Perm())}

			// Read content of file
			content, err := ioutil.ReadFile(FixPath(file))
			if err != nil {
				settings.ProgressError(err, relativePath)
				return
			}

			// Calc hash from content
			item.Hash = sha512.Sum512(content)

			// Find duplicated files by hash & size
			dup := findDuplicate(&item, &fat)
			if dup != nil {
				// Inc duplicated counters
				dupCount++
				dupSize += dup.Size
				item.Offset = dup.Offset
				item.Size = dup.Size
			} else {
				// Transform content of file (encrypt, compress)
				blob, err := TransformPack(content, header.Compress, header.Encrypt, shaKey)
				if err != nil {
					settings.ProgressError(err, relativePath)
					return
				}
				item.Size = int64(len(blob))
				fat.Size += item.Size

				// Write transformed content to package
				_, err = dest.Write(blob)
				if err != nil {
					settings.ProgressError(err, relativePath)
					return
				}
			}
			// Add new item to FAT
			fat.Items = append(fat.Items, item)
			fat.Count++

		}(i, file)
	}

	// Update progress state
	settings.Progress(fileCount, fileCount, "")

	// Encode FAT to JSON
	json, err := fat.JSON()
	if err != nil {
		return settings.FinishError(err)
	}

	// Transform FAT (encrypt, compress)
	fatBlob, err := TransformPack(json, header.Compress, header.Encrypt, shaKey)
	if err != nil {
		return settings.FinishError(err)
	}

	// Write FAT to package
	header.FatSize = int64(len(fatBlob))
	dest.Write(fatBlob)

	// Create a footer
	footer := NewFooter()
	footer.PackSize = HEADER_SIZE + fat.Size + header.FatSize + FOOTER_SIZE

	// Refresh FatSize in the header of package
	dest.Seek(0, os.SEEK_SET)
	err = header.Write(dest)
	if err != nil {
		return settings.FinishError(err)
	}

	// Write footer
	dest.Seek(0, os.SEEK_END)
	err = footer.Write(dest)
	if err != nil {
		return settings.FinishError(err)
	}

	// Process finished
	return settings.Finish(nil, fat.Count, footer.PackSize, dupCount, dupSize)
}
