package icepacker

import (
	"bytes"
	"fmt"

	"github.com/zhuangsirui/binpacker"

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

	b := make([]byte, FOOTER_SIZE)

	if _, err := io.ReadAtLeast(pack, b, len(b)); err != nil {
		return nil, err
	}

	unpacker := binpacker.NewUnpacker(bytes.NewBuffer(b))

	unpacker.FetchUint32(&footer.Checksum)
	unpacker.FetchInt64(&footer.PackSize)
	unpacker.FetchBytes(MAGIC_SIZE, &footer.Magic)

	err := unpacker.Error()
	if err != nil {
		return nil, err
	}

	// Validation

	// TODO: check checksum

	if !bytes.Equal(footer.Magic, []byte(MagicBytes)) {
		return nil, errors.New("Invalid file format!")
	}

	if footer.PackSize <= 0 {
		return nil, fmt.Errorf("Invalid pack size %d!", footer.PackSize)
	}

	return footer, nil
}

// Write writes the Footer struct to the io.Writer
func (footer *Footer) Write(writer io.Writer) error {

	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(buffer)
	packer.PushUint32(footer.Checksum)
	packer.PushInt64(footer.PackSize)
	packer.PushBytes(footer.Magic)

	err := packer.Error()
	if err != nil {
		return err
	}

	writer.Write(buffer.Bytes())

	return nil
}
