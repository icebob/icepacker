package icepacker

import (
	"crypto/sha512"
	"io/ioutil"
	"os"
	"path/filepath"
)

// BundleSettings records the settings of the bundle file
type BundleSettings struct {
	Compression byte
	Encryption  byte
	CipherKey   []byte
}

// BundleFile contains all info from bundle
type BundleFile struct {
	Path     string
	File     *os.File
	FAT      FAT
	Header   *Header
	Footer   *Footer
	DupCount int
	DupSize  int64
	Settings BundleSettings
	edited   bool
}

// findDuplication finds the duplicated file contents by hash of content.
func (this *BundleFile) findDuplicate(newItem *FATItem) *FATItem {
	for _, item := range this.FAT.Items {
		if newItem.Hash == item.Hash && newItem.OrigSize == item.OrigSize {
			return &item
		}
	}
	return nil
}

// CreateBundle created a new bundle file & struct.
func CreateBundle(filename string, settings BundleSettings) (*BundleFile, error) {

	// Create folders for target file
	err := os.MkdirAll(filepath.Dir(filename), DEFAULT_PERMISSION)
	if err != nil {
		return nil, err
	}

	// Create bundle file
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	// Create FAT
	fat := FAT{Count: 0, Size: 0}

	// Creat new Bundle
	bundle := BundleFile{Path: filename, File: f, FAT: fat, DupCount: 0, DupSize: 0, Settings: settings, edited: true}

	// Create a new header
	bundle.Header = NewHeader(settings.Encryption, settings.Compression)

	// Create a new footer
	bundle.Footer = NewFooter()

	err = bundle.Header.Write(bundle.File)
	if err != nil {
		return nil, err
	}

	return &bundle, nil
}

/*
func OpenBundle(filename string) (*BundleFile, error) {

}
*/

// AddFile adds a file to the bundle file
func (this *BundleFile) AddFile(relativePath, file string) (*FATItem, error) {

	// Open source file
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Get file info
	fileInfo, err := f.Stat()
	if err != nil {
		return nil, err
	}

	fileMode := fileInfo.Mode()
	fileSize := fileInfo.Size()

	// Create a FAT item
	item := FATItem{Path: filepath.ToSlash(relativePath), Offset: this.FAT.Size, OrigSize: fileSize, MTime: fileInfo.ModTime().UnixNano(), Mode: uint32(fileMode), Perm: uint32(fileMode.Perm())}

	// Read content of file
	content, err := ioutil.ReadFile(FixPath(file))
	if err != nil {
		return nil, err
	}

	// Calc hash from content
	item.Hash = sha512.Sum512(content)

	// Find duplicated files by hash & size
	dup := this.findDuplicate(&item)
	if dup != nil {
		// Inc duplicated counters
		this.DupCount++
		this.DupSize += dup.Size
		item.Offset = dup.Offset
		item.Size = dup.Size
	} else {
		// Transform content of file (encrypt, compress)
		blob, err := TransformPack(content, this.Settings.Compression, this.Settings.Encryption, this.Settings.CipherKey)
		if err != nil {
			return nil, err
		}
		item.Size = int64(len(blob))
		this.FAT.Size += item.Size

		// Write transformed content to package
		// TODO seek to the item.Offset before write (maybe read other file and position changed)
		_, err = this.File.Write(blob)
		if err != nil {
			return nil, err
		}
	}
	// Add new item to FAT
	this.FAT.Items = append(this.FAT.Items, item)
	this.FAT.Count++

	return &item, nil
}

/*
func (this *BundleFile) ReadFile(filepath string) ([]byte, error) {

}
*/

// Flush writes the footer of bundle
func (this *BundleFile) Flush() error {
	if this.edited {
		// Encode FAT to JSON
		json, err := this.FAT.JSON()
		if err != nil {
			return err
		}

		// Transform FAT (encrypt, compress)
		fatBlob, err := TransformPack(json, this.Settings.Compression, this.Settings.Encryption, this.Settings.CipherKey)
		if err != nil {
			return err
		}

		// Write FAT to package
		// TODO seek to the correct position before write (maybe read other file and position changed)
		this.Header.FatSize = int64(len(fatBlob))
		this.File.Write(fatBlob)

		// Set the PackSize in the footer
		this.Footer.PackSize = HEADER_SIZE + this.FAT.Size + this.Header.FatSize + FOOTER_SIZE

		// Refresh FatSize in the header of package
		this.File.Seek(0, os.SEEK_SET)
		err = this.Header.Write(this.File)
		if err != nil {
			return err
		}

		// Write footer
		this.File.Seek(0, os.SEEK_END)
		err = this.Footer.Write(this.File)
		if err != nil {
			return err
		}

	}

	return nil
}

// Close closes the bundle File
func (this *BundleFile) Close() error {
	if this.File != nil {
		this.File.Close()
		this.File = nil
	}
	return nil
}
