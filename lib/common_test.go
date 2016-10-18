package icepacker

import (
	"runtime"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFixPath(t *testing.T) {

	if runtime.GOOS != "windows" {
		t.Skip("Only on Windows")
		return
	}

	Convey("Modify path in Windows", t, func() {

		longPath := "foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz\\foo\\bar\\baz"

		var tests = []struct {
			value    string
			expected string
		}{
			{"index.js", "index.js"},
			{longPath, "\\\\?\\" + longPath},
			{"c:\\" + longPath, "\\\\?\\c:\\" + longPath},
			{"\\\\computer\\" + longPath, "\\\\?\\UNC\\computer\\" + longPath},
		}

		for _, test := range tests {
			So(FixPath(test.value), ShouldEqual, test.expected)
		}

	})

}
