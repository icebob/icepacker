package icepacker

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPacking(t *testing.T) {

	Convey("Should give error if the source is not exist", t, func() {

		source, _ := filepath.Abs("testdata/notvalid")
		target, _ := filepath.Abs("testdata/packed/notvalid.pack")

		result := Pack(PackSettings{
			SourceDir:      source,
			TargetFilename: target,
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

	Convey("Should make empty pack if the source is empty", t, func() {

		source, _ := filepath.Abs("testdata/empty")
		target, _ := filepath.Abs("testdata/packed/empty.pack")

		os.MkdirAll(source, DEFAULT_PERMISSION)

		result := Pack(PackSettings{
			SourceDir:      source,
			TargetFilename: target,
		})

		So(result, ShouldNotBeNil)
		So(result.Err, ShouldBeNil)
		So(result.FileCount, ShouldEqual, 0)
		So(result.Size, ShouldEqual, 74)
		So(result.DupCount, ShouldEqual, 0)
		So(result.DupSize, ShouldEqual, 0)

		os.Remove(target)
		os.RemoveAll(source)
	})

}
