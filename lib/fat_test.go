package icepacker

import (
	"crypto/sha512"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestToJson(t *testing.T) {

	fat := FAT{Count: 1, Size: 234}

	Convey("Give JSON string from one item", t, func() {

		fat.Items = append(fat.Items, FATItem{
			Path:     "filepath",
			Offset:   12345,
			Size:     88888,
			OrigSize: 343434,
			Hash:     sha512.Sum512([]byte("icepacker")),
			MTime:    123456789,
			Mode:     100,
			Perm:     777,
		})

		res, _ := fat.JSON()
		So(string(res), ShouldEqual, "{\"count\":1,\"size\":234,\"items\":[{\"path\":\"filepath\",\"offset\":12345,\"size\":88888,\"origSize\":343434,\"mTime\":123456789,\"mode\":100,\"perm\":777}]}")

	})

	Convey("Give JSON string from two item", t, func() {

		fat.Items = append(fat.Items, FATItem{
			Path:     "path1",
			Offset:   454545,
			Size:     2345,
			OrigSize: 556,
			Hash:     sha512.Sum512([]byte("item2")),
			MTime:    556456456,
			Mode:     130,
			Perm:     600,
		})

		res, _ := fat.JSON()
		So(string(res), ShouldEqual, "{\"count\":1,\"size\":234,\"items\":[{\"path\":\"filepath\",\"offset\":12345,\"size\":88888,\"origSize\":343434,\"mTime\":123456789,\"mode\":100,\"perm\":777},{\"path\":\"path1\",\"offset\":454545,\"size\":2345,\"origSize\":556,\"mTime\":556456456,\"mode\":130,\"perm\":600}]}")

	})

}

func TestFATFromJSON(t *testing.T) {

	Convey("Should give empty FAT struct from empty JSON string", t, func() {

		fat, err := FATFromJSON([]byte("{}"))
		So(err, ShouldBeNil)
		So(fat.Count, ShouldEqual, 0)
		So(fat.Size, ShouldEqual, 0)
		So(fat.Items, ShouldBeEmpty)
	})

	Convey("Should full FAT struct from JSON string", t, func() {

		fat, err := FATFromJSON([]byte("{\"count\":2,\"size\":760,\"items\":[{\"path\":\"file2\",\"offset\":231,\"size\":45853,\"mTime\":23123123,\"mode\":122,\"perm\":777},{\"path\":\"file1\",\"offset\":4343223,\"size\":5668,\"mTime\":6677886,\"mode\":454,\"perm\":332}]}"))
		So(err, ShouldBeNil)
		So(fat.Count, ShouldEqual, 2)
		So(fat.Size, ShouldEqual, 760)
		So(fat.Items, ShouldHaveLength, 2)

		item1 := fat.Items[0]
		So(item1.Path, ShouldEqual, "file2")
		So(item1.Offset, ShouldEqual, 231)
		So(item1.Size, ShouldEqual, 45853)
		So(item1.OrigSize, ShouldEqual, 0)
		//So(item1.Hash, ShouldEqual, 0)
		So(item1.MTime, ShouldEqual, 23123123)
		So(item1.Mode, ShouldEqual, 122)
		So(item1.Perm, ShouldEqual, 777)

		item2 := fat.Items[1]
		So(item2.Path, ShouldEqual, "file1")
		So(item2.Offset, ShouldEqual, 4343223)
		So(item2.Size, ShouldEqual, 5668)
		So(item2.OrigSize, ShouldEqual, 0)
		//So(item2.Hash, ShouldEqual, 0)
		So(item2.MTime, ShouldEqual, 6677886)
		So(item2.Mode, ShouldEqual, 454)
		So(item2.Perm, ShouldEqual, 332)
	})

	Convey("Should equal FromJSON <-> toJSON", t, func() {

		base := "{\"count\":2,\"size\":760,\"items\":[{\"path\":\"file2\",\"offset\":231,\"size\":45853,\"origSize\":0,\"mTime\":23123123,\"mode\":122,\"perm\":777},{\"path\":\"file1\",\"offset\":4343223,\"size\":5668,\"origSize\":0,\"mTime\":6677886,\"mode\":454,\"perm\":332}]}"
		fat, err := FATFromJSON([]byte(base))
		So(err, ShouldBeNil)
		json, err := fat.JSON()
		So(err, ShouldBeNil)
		So(string(json), ShouldEqual, base)
	})
}

func TestFATtoString(t *testing.T) {

	Convey("Should give string of FAT struct", t, func() {

		fat := FAT{Count: 1, Size: 234}
		So(fat.String(), ShouldEqual, "Count: 1, Size: 234")
	})

}

func TestFATItemtoString(t *testing.T) {

	Convey("Should give string of FATItem struct", t, func() {

		item := FATItem{
			Path:     "filepath",
			Offset:   12345,
			Size:     88888,
			OrigSize: 343434,
			Hash:     sha512.Sum512([]byte("icepacker")),
			MTime:    123456789,
			Mode:     100,
			Perm:     777,
		}

		So(item.String(), ShouldEqual, "path: filepath, offset: 12345, size: 88888, mode: 64d perm: 777")
	})

}
