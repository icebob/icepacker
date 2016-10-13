package icepacker

import (
	"errors"
	"os"
)

// ListSettings records settings of the listing
type ListSettings struct {
	PackFileName string
	Cipher       CipherSettings
	OnFinish     chan ListResult
}

// Finish returns a success ListResult instance and put to the OnFinish channel if it's not nil
func (this *ListSettings) Finish(err error, fat *FAT) ListResult {
	ret := ListResult{err, fat}
	if this.OnFinish != nil {
		this.OnFinish <- ret
	}
	return ret
}

// Finish returns an errored ListResult instance and put to the OnFinish channel if it's not nil
func (this *ListSettings) FinishError(err error) ListResult {
	ret := ListResult{Err: err}
	if this.OnFinish != nil {
		this.OnFinish <- ret
	}
	return ret
}

// ListPack lists the FAT from the package. Returns a ListResult instance with the FAT
func ListPack(settings ListSettings) ListResult {

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
	_, err = pack.Seek(-footer.PackSize, os.SEEK_END)
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

	// List finished
	return settings.Finish(nil, fat)
}
