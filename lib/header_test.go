package icepacker

import (
	"bytes"
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewHeaderConstructor(t *testing.T) {

	Convey("Should load default values", t, func() {

		header := NewHeader(ENCRYPT_AES, COMPRESS_GZIP)

		So(header.Magic, ShouldResemble, []byte(MagicBytes))
		So(header.Version, ShouldEqual, VERSION_1)
		So(header.Encrypt, ShouldEqual, ENCRYPT_AES)
		So(header.Compress, ShouldEqual, COMPRESS_GZIP)
		So(header.FatSize, ShouldEqual, 0)
		So(header.Created, ShouldBeLessThanOrEqualTo, time.Now().UnixNano())
	})

}

func TestHeaderWrite(t *testing.T) {

	Convey("Should write Header struct to Writer", t, func() {

		header := NewHeader(ENCRYPT_AES, COMPRESS_GZIP)
		header.FatSize = 12345
		header.Created = 123456789

		w := new(bytes.Buffer)
		err := header.Write(w)
		So(err, ShouldBeNil)
		So(w.Bytes(), ShouldResemble, []uint8{73, 80, 65, 67, 75, 1, 1, 1, 57, 48, 0, 0, 0, 0, 0, 0, 21, 205, 91, 7, 0, 0, 0, 0})
	})

}

func TestGetHeader(t *testing.T) {

	Convey("Should load Header struct from Reader", t, func() {
		r := bytes.NewReader([]uint8{73, 80, 65, 67, 75, 1, 1, 1, 57, 48, 0, 0, 0, 0, 0, 0, 21, 205, 91, 7, 0, 0, 0, 0})

		header, err := GetHeader(r)
		So(err, ShouldBeNil)
		So(header, ShouldNotBeNil)
		So(string(header.Magic), ShouldEqual, MagicBytes)
		So(header.Version, ShouldEqual, VERSION_1)
		So(header.Encrypt, ShouldEqual, ENCRYPT_AES)
		So(header.Compress, ShouldEqual, COMPRESS_GZIP)
		So(header.FatSize, ShouldEqual, 12345)
		So(header.Created, ShouldEqual, 123456789)
	})

	Convey("Should give error if size if small than HEADER_SIZE", t, func() {
		r := bytes.NewReader([]uint8{0, 0, 0, 0})

		header, err := GetHeader(r)
		So(err, ShouldResemble, errors.New("unexpected EOF"))
		So(header, ShouldBeNil)
	})

	Convey("Should give error if size Magic is not equal", t, func() {
		r := bytes.NewReader([]uint8{65, 80, 65, 67, 75, 1, 1, 1, 57, 48, 0, 0, 0, 0, 0, 0, 21, 205, 91, 7, 0, 0, 0, 0})

		header, err := GetHeader(r)
		So(err, ShouldResemble, errors.New("Invalid file format!"))
		So(header, ShouldBeNil)
	})

	Convey("Should give error if size Magic is not equal", t, func() {
		r := bytes.NewReader([]uint8{73, 80, 65, 67, 75, 2, 1, 1, 57, 48, 0, 0, 0, 0, 0, 0, 21, 205, 91, 7, 0, 0, 0, 0})

		header, err := GetHeader(r)
		So(err, ShouldResemble, errors.New("Invalid file version (2)!"))
		So(header, ShouldBeNil)
	})
}
