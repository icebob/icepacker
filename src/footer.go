package icepacker

import (
	"bufio"
	"bytes"
	"encoding/binary"

	"errors"

	"io"
)

// Footer is a struct for footer of package
type Footer struct {
	Checksum uint32
	PackSize int64
	Magic    []byte
}

// FOOTER_SIZE is the size of footer
const FOOTER_SIZE = 8 + 4 + MAGIC_SIZE

// NewFooter create a new Footer with default values
func NewFooter() *Footer {
	footer := new(Footer)

	footer.Checksum = 0
	footer.PackSize = 0
	footer.Magic = []byte(MagicBytes)

	return footer
}

// GetFooter reads the footer from the io.Reader
func GetFooter(pack io.Reader) (*Footer, error) {
	footer := NewFooter()

	r := bufio.NewReader(pack)

	if err := binary.Read(r, ByteOrder, &footer.Checksum); err != nil {
		return nil, err
	}

	if err := binary.Read(r, ByteOrder, &footer.PackSize); err != nil {
		return nil, err
	}

	if err := binary.Read(r, ByteOrder, &footer.Magic); err != nil {
		return nil, err
	}

	// TODO: check checksum

	if footer.PackSize <= 0 {
		return nil, errors.New("Invalid pack size!")
	}

	if !bytes.Equal(footer.Magic, []byte(MagicBytes)) {
		return nil, errors.New("Invalid file format!")
	}

	return footer, nil
}

// Write writes the Footer struct to the io.Writer
func (footer *Footer) Write(writer io.Writer) error {

	w := bufio.NewWriter(writer)
	if err := binary.Write(w, ByteOrder, footer.Checksum); err != nil {
		return err
	}

	if err := binary.Write(w, ByteOrder, footer.PackSize); err != nil {
		return err
	}

	if err := binary.Write(w, ByteOrder, footer.Magic); err != nil {
		return err
	}

	// Flush
	if err := w.Flush(); err != nil {
		return err
	}

	return nil
}
