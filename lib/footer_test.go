package icepacker

import (
	"bytes"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewFooterConstructor(t *testing.T) {

	Convey("Should load default values", t, func() {

		footer := NewFooter()

		So(footer.Checksum, ShouldEqual, 0)
		So(footer.PackSize, ShouldEqual, 0)
		So(footer.Magic, ShouldResemble, []byte(MagicBytes))
	})

}

func TestFooterWrite(t *testing.T) {

	Convey("Should write Footer struct to Writer", t, func() {

		footer := NewFooter()
		footer.PackSize = 12345
		footer.Checksum = 123456789

		w := new(bytes.Buffer)
		err := footer.Write(w)
		So(err, ShouldBeNil)
		So(w.Bytes(), ShouldResemble, []uint8{7, 91, 205, 21, 0, 0, 0, 0, 0, 0, 48, 57, 73, 80, 65, 67, 75})
	})

}

func TestGetFooter(t *testing.T) {

	Convey("Should load Footer struct from Reader", t, func() {
		r := bytes.NewReader([]uint8{7, 91, 205, 21, 0, 0, 0, 0, 0, 0, 48, 57, 73, 80, 65, 67, 75})

		footer, err := GetFooter(r)
		So(err, ShouldBeNil)
		So(footer, ShouldNotBeNil)
		So(string(footer.Magic), ShouldEqual, MagicBytes)
		So(footer.PackSize, ShouldEqual, 12345)
		So(footer.Checksum, ShouldEqual, 123456789)
	})

	Convey("Should give error if size if small than HEADER_SIZE", t, func() {
		r := bytes.NewReader([]uint8{0, 0})

		footer, err := GetFooter(r)
		So(err, ShouldResemble, errors.New("unexpected EOF"))
		So(footer, ShouldBeNil)
	})

	Convey("Should give error if size Magic is not equal", t, func() {
		r := bytes.NewReader([]uint8{7, 91, 205, 21, 0, 0, 0, 0, 0, 0, 48, 57, 73, 80, 65, 67, 72})

		footer, err := GetFooter(r)
		So(err, ShouldResemble, errors.New("Invalid file format!"))
		So(footer, ShouldBeNil)
	})

	Convey("Should give error if PackSize is negative", t, func() {
		r := bytes.NewReader([]uint8{7, 91, 205, 21, 255, 255, 255, 255, 255, 255, 255, 57, 73, 80, 65, 67, 75})

		footer, err := GetFooter(r)
		So(err, ShouldResemble, errors.New("Invalid pack size!"))
		So(footer, ShouldBeNil)
	})
}
