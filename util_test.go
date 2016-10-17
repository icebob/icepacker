package main

import (
	"runtime"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClearLine(t *testing.T) {

	Convey("Give a current line clear string", t, func() {

		actual := ClearLine()
		if runtime.GOOS == "windows" {
			So(actual, ShouldEqual, "                    \r")
		} else {
			So(actual, ShouldEqual, "\r")
		}

	})

}

func TestFormatBytes(t *testing.T) {

	Convey("Test formatter", t, func() {

		var tests = []struct {
			value    int64
			expected string
		}{
			{0, "0 B"},
			{10, "10 B"},
			{100, "100 B"},
			{1000, "1000 B"},
			{1024, "1.000 KiB"},
			{10000, "9.766 KiB"},
			{10 * 1024, "10.000 KiB"},
			{100 * 1024, "100.000 KiB"},
			{1024 * 1024, "1.000 MiB"},
			{1024 * 1024 * 1024, "1.000 GiB"},
			{1024 * 1024 * 1024 * 1024, "1.000 TiB"},
		}

		for _, test := range tests {
			So(FormatBytes(test.value), ShouldEqual, test.expected)
		}

	})

}

func TestFormatSeconds(t *testing.T) {

	Convey("Test formatter", t, func() {

		var tests = []struct {
			value    uint64
			expected string
		}{
			{0, "0:00"},
			{10, "0:10"},
			{100, "1:40"},
			{1000, "16:40"},
			{2000, "33:20"},
			{5000, "1:23:20"},
			{10 * 1000, "2:46:40"},
			{50 * 1000, "13:53:20"},
			{100 * 1000, "27:46:40"},
		}

		for _, test := range tests {
			So(FormatSeconds(test.value), ShouldEqual, test.expected)
		}

	})

}

func TestFormatPercent(t *testing.T) {

	Convey("Test formatter", t, func() {

		var tests = []struct {
			num      uint64
			denom    uint64
			expected string
		}{
			{0, 0, ""},
			{10, 0, ""},
			{100, 0, ""},

			{0, 1, "  0.00% [                    ]"},
			{0, 10, "  0.00% [                    ]"},

			{10, 100, " 10.00% [==                  ]"},
			{15, 100, " 15.00% [===                 ]"},
			{30, 100, " 30.00% [======              ]"},
			{50, 100, " 50.00% [==========          ]"},
			{90, 100, " 90.00% [==================  ]"},
			{100, 100, "100.00% [====================]"},
			{123, 100, "100.00% [====================]"},
		}

		for _, test := range tests {
			So(FormatPercent(test.num, test.denom), ShouldEqual, test.expected)
		}

	})

}

func TestRound(t *testing.T) {

	Convey("Test positive values", t, func() {

		var tests = []struct {
			value    float64
			expected int
		}{
			{0, 0},

			{0.4, 0},
			{0.49, 0},
			{0.5, 1},
			{0.55, 1},
			{0.6, 1},
		}

		for _, test := range tests {
			So(Round(test.value), ShouldEqual, test.expected)
		}

	})

	Convey("Test negative values", t, func() {

		var tests = []struct {
			value    float64
			expected int
		}{
			{-0.4, 0},
			{-0.49, 0},
			{-0.5, -1},
			{-0.55, -1},
			{-0.6, -1},
		}

		for _, test := range tests {
			So(Round(test.value), ShouldEqual, test.expected)
		}

	})

}
