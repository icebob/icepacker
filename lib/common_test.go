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

		var tests = []struct {
			value    string
			expected string
		}{
			{"index.js", "\\\\?\\index.js"},
			{"foo\\bar\\baz", "\\\\?\\foo\\bar\\baz"},
			{"c:\\foo\\bar\\baz", "\\\\?\\c:\\foo\\bar\\baz"},
			{"\\\\computer\\foo\\bar", "\\\\?\\UNC\\computer\\foo\\bar"},
		}

		for _, test := range tests {
			So(FixPath(test.value), ShouldEqual, test.expected)
		}

	})

}
