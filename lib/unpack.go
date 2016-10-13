package icepacker

import (
	"errors"
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

	// Create target directory
	err := os.MkdirAll(settings.TargetDir, DEFAULT_PERMISSION)
	if err != nil {
		return settings.FinishError(err)
	}

	// Hash the cipher key
	shaKey := HashingKey(settings.Cipher)

	// Open package file
	pack, err := os.Open(settings.PackFileName)
	if err != nil {
		return settings.FinishError(err)
	}
	defer pack.Close()

	// Get package file info
	packFileInfo, err := pack.Stat()
	if err != nil {
		return settings.FinishError(err)
	}

	// Check the size of package (minimum HEADER_SIZE + FOOTER_SIZE)
	size := packFileInfo.Size()
	if size < HEADER_SIZE+FOOTER_SIZE {
		return settings.FinishError(errors.New("File is too small!"))
	}

	// 1. Jump to end of file
	_, err = pack.Seek(-FOOTER_SIZE, os.SEEK_END)
	if err != nil {
		return settings.FinishError(err)
	}

	// Read file footer
	footer, err := GetFooter(pack)
	if err != nil {
		return settings.FinishError(err)
	}

	// 2. Ha stimmel a Magic, akkor a fájl elejére ugrani PackSize alapján (lehet, hogy mögé másolt)
	fileBegin, err := pack.Seek(-footer.PackSize, os.SEEK_END)
	if err != nil {
		return settings.FinishError(err)
	}

	// 3. Read file header
	header, err := GetHeader(pack)

	// 4. jump to FAT
	_, err = pack.Seek(-(FOOTER_SIZE + header.FatSize), os.SEEK_END)
	if err != nil {
		return settings.FinishError(err)
	}

	// 5. Read FAT
	fatBuf := make([]byte, header.FatSize)
	_, err = pack.Read(fatBuf)
	if err != nil {
		return settings.FinishError(err)
	}

	// Transform back the FAT (decompress, decrypt)
	fatContent, err := TransformUnpack(fatBuf, header.Compress, header.Encrypt, shaKey)
	if err != nil {
		return settings.FinishError(err)
	}

	fat, err := FATFromJSON(fatContent)
	if err != nil {
		return settings.FinishError(err)
	}

	// 6. Restore files from package
	totalSize := int64(0)
	dataBaseOffset := fileBegin + HEADER_SIZE
	fileCount := len(fat.Items)
	for i, item := range fat.Items {

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
				_, err = pack.Seek(dataBaseOffset+item.Offset, os.SEEK_SET)
				if err != nil {
					settings.ProgressError(err, item.Path)
					return
				}

				// Read the content
				blob := make([]byte, item.Size)
				_, err = pack.Read(blob)
				if err != nil {
					settings.ProgressError(err, item.Path)
					return
				}

				// Transform back (decompress, decrypt)
				content, err := TransformUnpack(blob, header.Compress, header.Encrypt, shaKey)
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
