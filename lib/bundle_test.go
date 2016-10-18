package icepacker

import (
	"crypto/sha512"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFindDuplicate(t *testing.T) {

	bundle := BundleFile{FAT: FAT{Count: 3}}

	Convey("Should give duplicated item", t, func() {

		bundle.FAT.Items = append(bundle.FAT.Items, FATItem{
			Path:     "file1.txt",
			OrigSize: 12300,
			Hash:     sha512.Sum512([]byte("other")),
		})

		bundle.FAT.Items = append(bundle.FAT.Items, FATItem{
			Path:     "file88.png",
			OrigSize: 800,
			Hash:     sha512.Sum512([]byte("icepacker")),
		})

		bundle.FAT.Items = append(bundle.FAT.Items, FATItem{
			Path:     "data.log",
			OrigSize: 12300,
			Hash:     sha512.Sum512([]byte("icepacker")),
		})

		dup := bundle.FindDuplicate(&FATItem{
			Path:     "foo.bar",
			OrigSize: 12300,
			Hash:     sha512.Sum512([]byte("icepacker")),
		})

		So(dup, ShouldNotBeNil)
		So(dup.Path, ShouldEqual, "data.log")
		So(dup.OrigSize, ShouldEqual, 12300)
	})
	/*
	   func TestCreateBundle(t *testing.T) {
	       // Check FAT, Header, Footer, DataBaseOffset
	   }

	   func TestOpenBundle(t *testing.T) {
	       // Check FAT, Header, Footer, DataBaseOffset
	   }

	   func TestAddFile(t *testing.T) {

	   }

	   func TestReadFileFromPath(t *testing.T) {

	   }

	   func TestReadFile(t *testing.T) {

	   }

	   func TestFinalize(t *testing.T) {

	   }

	   func TestClose(t *testing.T) {

	   }
	*/
}
