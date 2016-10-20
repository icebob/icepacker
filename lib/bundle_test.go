package icepacker

import (
	"crypto/sha512"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var bundlePath, err = filepath.Abs("testdata/bundle/bundle.pack")
var key = HashingKey(CipherSettings{Key: "password", Iteration: 10000})
var bundle *BundleFile

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
}

func TestCreateBundle(t *testing.T) {

	Convey("create a new bundle file", t, func() {

		os.Remove(bundlePath)
		bundle, err = CreateBundle(bundlePath, BundleSettings{
			Compression: COMPRESS_NONE,
			Encryption:  ENCRYPT_NONE,
			CipherKey:   key,
		})

		So(err, ShouldBeNil)
		So(bundle, ShouldNotBeNil)

		So(bundle.Path, ShouldEqual, bundlePath)
		So(bundle.DataBaseOffset, ShouldEqual, HEADER_SIZE)

		So(bundle.Header, ShouldNotBeNil)
		So(bundle.Header.Compress, ShouldEqual, COMPRESS_NONE)
		So(bundle.Header.Encrypt, ShouldEqual, ENCRYPT_NONE)

		So(bundle.Footer, ShouldNotBeNil)

		So(bundle.FAT, ShouldNotBeNil)
		So(bundle.FAT.Count, ShouldEqual, 0)
		So(bundle.FAT.Size, ShouldEqual, 0)
	})

	Convey("add a file to bundle", t, func() {
		filename, _ := filepath.Abs("testdata/simple/file1.txt")
		item, err := bundle.AddFile("file1.txt", filename)

		So(err, ShouldBeNil)
		So(item, ShouldNotBeNil)

		// Check item
		So(item.Path, ShouldEqual, "file1.txt")
		So(item.Offset, ShouldEqual, 0)
		So(item.OrigSize, ShouldEqual, 5)
		So(item.Size, ShouldEqual, 5)

		// Check bundle
		So(bundle.FAT, ShouldNotBeNil)
		So(bundle.FAT.Count, ShouldEqual, 1)
		So(bundle.FAT.Size, ShouldEqual, 5)

	})

	Convey("finalize the bundle", t, func() {
		err := bundle.Finalize()
		So(err, ShouldBeNil)

		So(bundle.Footer, ShouldNotBeNil)
		So(bundle.Footer.PackSize, ShouldEqual, 180)
		So(bundle.Header.FatSize, ShouldEqual, 134)

		// Close the bundle
		err = bundle.Close()
		So(err, ShouldBeNil)
	})
}

func TestOpenBundle(t *testing.T) {

	Convey("open the bundle file", t, func() {

		bundle, err = OpenBundle(bundlePath, key)

		So(err, ShouldBeNil)
		So(bundle, ShouldNotBeNil)

		So(bundle.Path, ShouldEqual, bundlePath)
		So(bundle.DataBaseOffset, ShouldEqual, HEADER_SIZE)

		So(bundle.Header, ShouldNotBeNil)
		So(bundle.Header.Compress, ShouldEqual, COMPRESS_NONE)
		So(bundle.Header.Encrypt, ShouldEqual, ENCRYPT_NONE)

		So(bundle.Footer, ShouldNotBeNil)

		So(bundle.FAT, ShouldNotBeNil)
		So(bundle.FAT.Count, ShouldEqual, 1)
		So(bundle.FAT.Size, ShouldEqual, 5)
	})

	Convey("add other file to an exist bundle", t, func() {
		filename, _ := filepath.Abs("testdata/simple/file2.txt")
		item, err := bundle.AddFile("file2.txt", filename)

		So(err, ShouldBeNil)
		So(item, ShouldNotBeNil)

		// Check item
		So(item.Path, ShouldEqual, "file2.txt")
		So(item.Offset, ShouldEqual, 5)
		So(item.OrigSize, ShouldEqual, 14)
		So(item.Size, ShouldEqual, 14)

		// Check bundle
		So(bundle.FAT.Count, ShouldEqual, 2)
		So(bundle.FAT.Size, ShouldEqual, 19)
	})

	Convey("add more files to an exist bundle", t, func() {
		filename, _ := filepath.Abs("testdata/simple/dir1/file3.txt")
		item, err := bundle.AddFile("dir1/file3.txt", filename)

		So(err, ShouldBeNil)
		So(item, ShouldNotBeNil)
		So(item.Offset, ShouldEqual, 19)

		// Check bundle
		So(bundle.FAT.Count, ShouldEqual, 3)
		So(bundle.FAT.Size, ShouldEqual, 2401)

		filename, _ = filepath.Abs("testdata/simple/dir1/icon1.png")
		item, err = bundle.AddFile("dir1/icon1.png", filename)

		So(err, ShouldBeNil)
		So(item, ShouldNotBeNil)
		So(item.Offset, ShouldEqual, 2401)

		// Check bundle
		So(bundle.FAT.Count, ShouldEqual, 4)
		So(bundle.FAT.Size, ShouldEqual, 3176)
		So(bundle.DupCount, ShouldEqual, 0)
		So(bundle.DupSize, ShouldEqual, 0)

	})

	Convey("add duplicated file to an exist bundle", t, func() {
		filename, _ := filepath.Abs("testdata/simple/dir2/icon-same.png")
		item, err := bundle.AddFile("dir2/icon-same.png", filename)

		So(err, ShouldBeNil)
		So(item, ShouldNotBeNil)
		So(item.Offset, ShouldEqual, 2401)

		// Check bundle
		So(bundle.FAT.Count, ShouldEqual, 5)
		So(bundle.FAT.Size, ShouldEqual, 3176)
		So(bundle.DupCount, ShouldEqual, 1)
		So(bundle.DupSize, ShouldEqual, 775)

	})

	Convey("finalize the bundle", t, func() {
		err := bundle.Finalize()
		So(err, ShouldBeNil)

		So(bundle.Footer, ShouldNotBeNil)
		So(bundle.Footer.PackSize, ShouldEqual, 3812)
		So(bundle.Header.FatSize, ShouldEqual, 595)

		// Close the bundle
		err = bundle.Close()
		So(err, ShouldBeNil)
	})
}

func TestReadFile(t *testing.T) {

	Convey("open the bundle file", t, func() {

		bundle, err = OpenBundle(bundlePath, key)

		So(err, ShouldBeNil)
		So(bundle, ShouldNotBeNil)
	})

	Convey("read a file by path", t, func() {
		content, err := bundle.ReadFileFromPath("dir1/file3.txt")
		So(err, ShouldBeNil)
		So(content, ShouldNotBeNil)
		So(len(content), ShouldEqual, 2382)

		// Check the content
		origContent, err := ioutil.ReadFile(FixPath("testdata/simple/dir1/file3.txt"))
		So(content, ShouldResemble, origContent)
	})

	Convey("read a non exists file by path", t, func() {
		content, err := bundle.ReadFileFromPath("dir1111/file123.txt")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "File not found")
		So(content, ShouldBeNil)
	})

	Convey("read a file by item", t, func() {
		content, err := bundle.ReadFile(bundle.FAT.Items[1])
		So(err, ShouldBeNil)
		So(content, ShouldNotBeNil)
		So(len(content), ShouldEqual, 14)

		// Check the content
		origContent, err := ioutil.ReadFile(FixPath("testdata/simple/file2.txt"))
		So(content, ShouldResemble, origContent)
	})

	Convey("close the bundle and delete", t, func() {
		// Close the bundle
		err = bundle.Close()
		So(err, ShouldBeNil)

		os.Remove(bundlePath)
	})
}
