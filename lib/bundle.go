package icepacker

import (
	"crypto/sha512"
	"errors"
	"fmt"
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
	Path           string
	File           *os.File
	FAT            FAT
	Header         *Header
	Footer         *Footer
	DataBaseOffset int64
	Settings       BundleSettings
	DupCount       int
	DupSize        int64
	edited         bool
}

// FindDuplication finds the duplicated file contents by hash of content.
func (this *BundleFile) FindDuplicate(newItem *FATItem) *FATItem {
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
	bundle := BundleFile{Path: filename, File: f, FAT: fat, Settings: settings, edited: true}

	// Create a new header
	bundle.Header = NewHeader(settings.Encryption, settings.Compression)

	// Set base offset of data block
	bundle.DataBaseOffset = HEADER_SIZE

	// Create a new footer
	bundle.Footer = NewFooter()

	err = bundle.Header.Write(bundle.File)
	if err != nil {
		return nil, err
	}

	return &bundle, nil
}

// OpenBundle open an exist bundle file. Load header, footer and FAT
func OpenBundle(filename string, cipherKey []byte) (*BundleFile, error) {

	var f *os.File
	var err error

	if filename == "-" {
		f = os.Stdin
	} else {
		// Open package file
		f, err = os.OpenFile(filename, os.O_RDWR, 0666)
		if err != nil {
			return nil, err
		}
	}

	// Get package file info
	packFileInfo, err := f.Stat()
	if err != nil {
		return nil, err
	}

	// Check the size of package (minimum HEADER_SIZE + FOOTER_SIZE)
	size := packFileInfo.Size()
	if size < HEADER_SIZE+FOOTER_SIZE {
		return nil, fmt.Errorf("File is too small! Size: %d", size)
	}

	// 1. Jump to end of file
	_, err = f.Seek(-FOOTER_SIZE, os.SEEK_END)
	if err != nil {
		return nil, err
	}

	// Read file footer
	footer, err := GetFooter(f)
	if err != nil {
		return nil, err
	}

	// Jump to the begin of bundle by PackSize (maybe bundle is behind other file)
	fileBegin, err := f.Seek(-footer.PackSize, os.SEEK_END)
	if err != nil {
		return nil, err
	}

	// Calc base offset of data block
	dataBaseOffset := fileBegin + HEADER_SIZE

	// 3. Read file header
	header, err := GetHeader(f)
	if err != nil {
		return nil, err
	}

	// 4. jump to FAT
	_, err = f.Seek(-(FOOTER_SIZE + header.FatSize), os.SEEK_END)
	if err != nil {
		return nil, err
	}

	// 5. Read FAT
	fatBuf := make([]byte, header.FatSize)
	_, err = f.Read(fatBuf)
	if err != nil {
		return nil, err
	}

	// Transform back the FAT (decompress, decrypt)
	fatContent, err := TransformUnpack(fatBuf, header.Compress, header.Encrypt, cipherKey)
	if err != nil {
		return nil, err
	}

	// Recover the FAT struct from JSON
	fat, err := FATFromJSON(fatContent)
	if err != nil {
		return nil, err
	}

	settings := BundleSettings{Compression: header.Compress, Encryption: header.Encrypt, CipherKey: cipherKey}
	bundle := BundleFile{Path: filename, File: f, FAT: *fat, Header: header, Footer: footer, Settings: settings, DataBaseOffset: dataBaseOffset}

	return &bundle, nil
}

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
	dup := this.FindDuplicate(&item)
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

		// jump to the end of the data block
		_, err = this.File.Seek(this.DataBaseOffset+item.Offset, os.SEEK_SET)
		if err != nil {
			return nil, err
		}

		// Write transformed content to package
		_, err = this.File.Write(blob)
		if err != nil {
			return nil, err
		}
	}
	// Add new item to FAT
	this.FAT.Items = append(this.FAT.Items, item)
	this.FAT.Count++

	this.edited = true

	return &item, nil
}

// ReadFileFromPath searches the FATItem in FAT by `filepath`` and reads
// the content of the file from the bundle
func (this *BundleFile) ReadFileFromPath(filepath string) ([]byte, error) {
	for _, item := range this.FAT.Items {
		if item.Path == filepath {
			return this.ReadFile(item)
		}
	}
	return nil, errors.New("File not found! Path: " + filepath)
}

// ReadFile reads the content of the file from the bundle
func (this *BundleFile) ReadFile(item FATItem) ([]byte, error) {

	// Seek to content
	_, err := this.File.Seek(this.DataBaseOffset+item.Offset, os.SEEK_SET)
	if err != nil {
		return nil, err
	}

	// Read the content
	blob := make([]byte, item.Size)
	_, err = this.File.Read(blob)
	if err != nil {
		return nil, err
	}

	// Transform back (decompress, decrypt)
	content, err := TransformUnpack(blob, this.Settings.Compression, this.Settings.Encryption, this.Settings.CipherKey)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// Finalize writes the footer of bundle
func (this *BundleFile) Finalize() error {
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

		// jump to the end of the data block
		_, err = this.File.Seek(this.DataBaseOffset+this.FAT.Size, os.SEEK_SET)
		if err != nil {
			return err
		}

		// Write FAT to package
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
