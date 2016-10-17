package icepacker

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/zhuangsirui/binpacker"
)

// Header is the header of package
type Header struct {
	Magic    []byte
	Version  byte
	Encrypt  byte
	Compress byte
	FatSize  int64
	Created  int64
}

// HEADER_SIZE is the size of the Header
const HEADER_SIZE = MAGIC_SIZE + 1 + 1 + 1 + 8 + 8

// NewHeader create a new Header with default values and set
// the encryption and compress types.
func NewHeader(encryption, compression byte) *Header {
	header := new(Header)

	header.Magic = []byte(MagicBytes)
	header.Version = VERSION_1
	header.Encrypt = encryption
	header.Compress = compression
	header.FatSize = 0
	header.Created = time.Now().UnixNano()

	return header
}

// GetHeader reads the header from the io.Reader and returns a *Header struct
func GetHeader(pack io.Reader) (*Header, error) {
	header := NewHeader(ENCRYPT_NONE, COMPRESS_NONE)

	b := make([]byte, HEADER_SIZE)

	if _, err := io.ReadAtLeast(pack, b, len(b)); err != nil {
		return nil, err
	}

	unpacker := binpacker.NewUnpacker(bytes.NewBuffer(b))

	unpacker.FetchBytes(MAGIC_SIZE, &header.Magic)
	unpacker.FetchByte(&header.Version)
	unpacker.FetchByte(&header.Encrypt)
	unpacker.FetchByte(&header.Compress)
	unpacker.FetchInt64(&header.FatSize)
	unpacker.FetchInt64(&header.Created)

	err := unpacker.Error()
	if err != nil {
		return nil, err
	}

	// Validation

	if !bytes.Equal(header.Magic, []byte(MagicBytes)) {
		return nil, errors.New("Invalid file format!")
	}

	if header.Version != VERSION_1 {
		return nil, fmt.Errorf("Invalid file version (%d)!", header.Version)
	}

	return header, nil
}

// Write writes the header to the io.Writer
func (header *Header) Write(writer io.Writer) error {

	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(buffer)
	packer.PushBytes(header.Magic)
	packer.PushByte(header.Version)
	packer.PushByte(header.Encrypt)
	packer.PushByte(header.Compress)
	packer.PushInt64(header.FatSize)
	packer.PushInt64(header.Created)

	err := packer.Error()
	if err != nil {
		return err
	}

	writer.Write(buffer.Bytes())

	return nil
}
