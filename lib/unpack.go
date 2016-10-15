package icepacker

import (
	"os"
	"path/filepath"
)

// UnpackSettings records the settings of the unpacking
type UnpackSettings struct {
	PackFileName string
	TargetDir    string
	Includes     string
	Excludes     string
	Cipher       CipherSettings
	OnProgress   chan ProgressState
	OnFinish     chan FinishResult
}

// Progress push a success ProgressState instance to the OnProgress channel.
func (this *UnpackSettings) Progress(total, index int, filename string) {
	if this.OnProgress != nil {
		this.OnProgress <- ProgressState{nil, total, index, filename}
	}
}

// ProgressError push an error ProgressState instance to the OnProgress channel.
func (this *UnpackSettings) ProgressError(err error, filename string) {
	if this.OnProgress != nil {
		this.OnProgress <- ProgressState{err, 0, 0, filename}
	}
}

// Finish returns a success FinishResult instance and put to the OnFinish channel if it's not nil
func (this *UnpackSettings) Finish(err error, fileCount int64, size int64, dupCount int, dupSize int64) FinishResult {
	ret := FinishResult{err, fileCount, size, dupCount, dupSize}
	if this.OnFinish != nil {
		this.OnFinish <- ret
	}
	return ret
}

// Finish returns an errored FinishResult instance and put to the OnFinish channel if it's not nil
func (this *UnpackSettings) FinishError(err error) FinishResult {
	ret := FinishResult{Err: err}
	if this.OnFinish != nil {
		this.OnFinish <- ret
	}
	return ret
}

// Unpack extract files from the package file
func Unpack(settings UnpackSettings) FinishResult {

	// Hash the cipher key
	shaKey := HashingKey(settings.Cipher)

	// Create target directory
	err := os.MkdirAll(settings.TargetDir, DEFAULT_PERMISSION)
	if err != nil {
		return settings.FinishError(err)
	}

	bundle, err := OpenBundle(settings.PackFileName, shaKey)
	if err != nil {
		return settings.FinishError(err)
	}
	defer bundle.Close()

	// 6. Restore files from package
	totalSize := int64(0)
	fileCount := len(bundle.FAT.Items)
	for i, item := range bundle.FAT.Items {

		func(i int, item FATItem) {

			fullPath := filepath.Join(settings.TargetDir, filepath.FromSlash(item.Path))
			dir := FixPath(filepath.Dir(fullPath))

			// Create directories by fullPath
			err = os.MkdirAll(dir, DEFAULT_PERMISSION)
			if err != nil {
				settings.ProgressError(err, item.Path)
				return
			}

			// Update progress state
			if i%100 == 0 {
				settings.Progress(fileCount, i, item.Path)
			}

			// Create new target file
			target, err := os.OpenFile(FixPath(fullPath), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(item.Perm))
			if err != nil {
				settings.ProgressError(err, item.Path)
				return
			}
			defer target.Close()

			// If not empty
			if item.Size > 0 {
				// Seek to content
				_, err = bundle.File.Seek(bundle.DataBaseOffset+item.Offset, os.SEEK_SET)
				if err != nil {
					settings.ProgressError(err, item.Path)
					return
				}

				// Read the content
				blob := make([]byte, item.Size)
				_, err = bundle.File.Read(blob)
				if err != nil {
					settings.ProgressError(err, item.Path)
					return
				}

				// Transform back (decompress, decrypt)
				content, err := TransformUnpack(blob, bundle.Settings.Compression, bundle.Settings.Encryption, shaKey)
				if err != nil {
					settings.ProgressError(err, item.Path)
					return
				}

				// Write the content to the target file
				_, err = target.Write(content)
				if err != nil {
					settings.ProgressError(err, item.Path)
					return
				}

				totalSize += int64(len(content))
			}
		}(i, item)

	}

	// Update progress to 100%
	settings.Progress(fileCount, fileCount, "")

	// Process finished
	return settings.Finish(nil, int64(fileCount), totalSize, 0, 0)
}
