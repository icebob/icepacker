package icepacker

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnpacking(t *testing.T) {

	Convey("Should give error if the source is not exist", t, func() {

		source, _ := filepath.Abs("testdata/packed/notexists.pack")
		target, _ := filepath.Abs("testdata/unpacked/notexists")

		result := Unpack(UnpackSettings{
			PackFileName: source,
			TargetDir:    target,
		})

		So(result, ShouldNotBeNil)
		So(result.Err, ShouldNotBeNil)
		// SKIP: Give different error message on different OS
		// So(result.Err.Error(), ShouldContainSubstring, "The system cannot find the file specified.")

		stat, err := os.Stat(target)
		So(stat, ShouldNotBeNil)
		So(err, ShouldBeNil)

		os.Remove(target)
	})

	Convey("Should give error if the source is not a package file", t, func() {

		source, _ := filepath.Abs("testdata/not-a.pack")
		target, _ := filepath.Abs("testdata/unpacked/not-a-pack")

		result := Unpack(UnpackSettings{
			PackFileName: source,
			TargetDir:    target,
		})

		So(result, ShouldNotBeNil)
		So(result.Err, ShouldNotBeNil)
		So(result.Err.Error(), ShouldEqual, "Invalid file format!")

		os.Remove(target)
	})

}
