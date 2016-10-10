package icepacker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
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

	r := bufio.NewReader(pack)

	if err := binary.Read(r, ByteOrder, &header.Magic); err != nil {
		return nil, err
	}

	var err error
	header.Version, err = r.ReadByte()
	if err != nil {
		return nil, err
	}

	header.Encrypt, err = r.ReadByte()
	if err != nil {
		return nil, err
	}

	header.Compress, err = r.ReadByte()
	if err != nil {
		return nil, err
	}

	if err := binary.Read(r, ByteOrder, &header.FatSize); err != nil {
		return nil, err
	}

	if err := binary.Read(r, ByteOrder, &header.Created); err != nil {
		return nil, err
	}

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

	w := bufio.NewWriter(writer)
	if err := binary.Write(w, ByteOrder, header.Magic); err != nil {
		return err
	}

	w.WriteByte(header.Version)
	w.WriteByte(header.Encrypt)
	w.WriteByte(header.Compress)

	if err := binary.Write(w, ByteOrder, header.FatSize); err != nil {
		return err
	}

	if err := binary.Write(w, ByteOrder, header.Created); err != nil {
		return err
	}

	// Flush
	if err := w.Flush(); err != nil {
		return err
	}

	return nil
}
