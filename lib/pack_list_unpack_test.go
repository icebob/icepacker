package icepacker

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPackListUnpack(t *testing.T) {

	Convey("Should pack & unpack directory", t, func() {

		var tests = []struct {
			compress  byte
			encrypt   byte
			fileCount int64
			size      int64
			dupCount  int
			dupSize   int64

			includes string
			excludes string
		}{
			{COMPRESS_NONE, ENCRYPT_NONE, 8, 4341, 1, 775, "", ""},
			{COMPRESS_NONE, ENCRYPT_AES, 8, 4456, 1, 791, "", ""},
			{COMPRESS_GZIP, ENCRYPT_NONE, 8, 2468, 1, 808, "", ""},
			{COMPRESS_GZIP, ENCRYPT_AES, 8, 2577, 1, 824, "", ""},

			// Test includes
			{COMPRESS_NONE, ENCRYPT_NONE, 4, 2913, 0, 0, ".txt$", ""},
			{COMPRESS_NONE, ENCRYPT_NONE, 2, 1078, 1, 775, ".png$", ""},
			{COMPRESS_NONE, ENCRYPT_NONE, 0, 74, 0, 0, ".pdf$", ""},
			{COMPRESS_NONE, ENCRYPT_NONE, 2, 3462, 0, 0, "dir1", ""},

			// Test excludes
			{COMPRESS_NONE, ENCRYPT_NONE, 4, 1494, 1, 775, "", ".txt$"},
			{COMPRESS_NONE, ENCRYPT_NONE, 6, 3330, 0, 0, "", ".png$"},
			{COMPRESS_NONE, ENCRYPT_NONE, 8, 4341, 1, 775, "", ".dat$"},
			{COMPRESS_NONE, ENCRYPT_NONE, 6, 3912, 0, 0, "", "dir2"},
		}

		for i, test := range tests {

			source, _ := filepath.Abs("testdata/simple")
			target, _ := filepath.Abs(fmt.Sprintf("testdata/packed/simple-%d.pack", i))
			unTarget, _ := filepath.Abs(fmt.Sprintf("testdata/unpacked/simple-%d", i))

			os.Remove(target)
			os.Remove(unTarget)

			// --- TEST PACKING ---
			result := Pack(PackSettings{
				SourceDir:      source,
				TargetFilename: target,
				Compression:    test.compress,
				Encryption:     test.encrypt,
				Cipher:         NewCipherSettings("PackSecretKey"),
				Includes:       test.includes,
				Excludes:       test.excludes,
			})

			// --- CHECK RESULT OF PACK
			So(result, ShouldNotBeNil)
			So(result.Err, ShouldBeNil)
			So(result.FileCount, ShouldEqual, test.fileCount)
			So(result.DupCount, ShouldEqual, test.dupCount)
			So(result.DupSize, ShouldEqual, test.dupSize)

			// Skip below asserts because it can be different, if compression enabled
			if test.compress == COMPRESS_NONE {
				So(result.Size, ShouldEqual, test.size)

				stat, err := os.Stat(target)
				So(err, ShouldBeNil)
				So(stat.Size(), ShouldEqual, test.size)
			}

			// --- TEST LISTING
			result2 := ListPack(ListSettings{
				PackFileName: target,
				Cipher:       NewCipherSettings("PackSecretKey"),
			})

			So(result2, ShouldNotBeNil)
			So(result2.Err, ShouldBeNil)
			So(result2.FAT, ShouldNotBeNil)
			So(result2.FAT.Count, ShouldEqual, test.fileCount)
			So(result2.FAT.Items, ShouldHaveLength, test.fileCount)

			// --- TEST UNPACKING
			result3 := Unpack(UnpackSettings{
				PackFileName: target,
				TargetDir:    unTarget,
				Cipher:       NewCipherSettings("PackSecretKey"),
				//Includes:       test.includes,
				//Excludes:       test.excludes,
			})

			So(result3, ShouldNotBeNil)
			So(result3.Err, ShouldBeNil)
			So(result3.FileCount, ShouldEqual, test.fileCount)

			// Clear
			os.Remove(target)
			os.RemoveAll(unTarget)
		}

	})

	Convey("Should pack & unpack only one file", t, func() {

		source, _ := filepath.Abs("testdata/onlyfile.txt")
		target, _ := filepath.Abs("testdata/packed/onlyfile.pack")
		unTarget, _ := filepath.Abs("testdata/unpacked")

		os.Remove(target)
		os.Remove(unTarget)

		// --- TEST PACKING ---
		result := Pack(PackSettings{
			SourceDir:      source,
			TargetFilename: target,
		})

		// --- CHECK RESULT OF PACK
		So(result, ShouldNotBeNil)
		So(result.Err, ShouldBeNil)
		So(result.FileCount, ShouldEqual, 1)
		So(result.DupCount, ShouldEqual, 0)
		So(result.DupSize, ShouldEqual, 0)

		// Skip below asserts because it can be different, if compression enabled
		/*
			So(result.Size, ShouldEqual, 18793)

			stat, err := os.Stat(target)
			So(err, ShouldBeNil)
			So(stat.Size(), ShouldEqual, 18793)
		*/

		// --- TEST LISTING
		result2 := ListPack(ListSettings{
			PackFileName: target,
			Cipher:       NewCipherSettings("PackSecretKey"),
		})

		So(result2, ShouldNotBeNil)
		So(result2.Err, ShouldBeNil)
		So(result2.FAT, ShouldNotBeNil)
		So(result2.FAT.Count, ShouldEqual, 1)
		So(result2.FAT.Items, ShouldHaveLength, 1)

		// --- TEST UNPACKING
		result3 := Unpack(UnpackSettings{
			PackFileName: target,
			TargetDir:    unTarget,
		})

		So(result3, ShouldNotBeNil)
		So(result3.Err, ShouldBeNil)
		So(result3.FileCount, ShouldEqual, 1)

		sourceStat, err := os.Stat(source)
		So(err, ShouldBeNil)
		unTargetStat, err := os.Stat(unTarget + "/onlyfile.txt")
		So(err, ShouldBeNil)
		So(unTargetStat.Size(), ShouldEqual, sourceStat.Size())

		// Clear
		os.Remove(target)
		os.Remove(unTarget + "/onlyfile.txt")

	})

}
