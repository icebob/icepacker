package icepacker

import (
	"encoding/binary"
	"runtime"
	"strings"
)

const DEFAULT_PERMISSION = 0755

// Magic bytes to identify the file format
const MagicBytes = "IPACK"

const MAGIC_SIZE = 5

const VERSION_1 = 1

var ByteOrder = binary.BigEndian

// Encryption enum constants
const (
	ENCRYPT_NONE = iota
	ENCRYPT_AES
)

// Compression enum constants
const (
	COMPRESS_NONE = iota
	COMPRESS_GZIP
)

// Fixing filepath on Windows to support longer filepath than 255 bytes.
// More information: https://msdn.microsoft.com/en-us/library/aa365247(VS.85).aspx
func FixPath(path string) string {
	const Prefix = `\\?\`

	if runtime.GOOS == "windows" {

		if !strings.HasPrefix(path, Prefix) {
			if strings.HasPrefix(path, `\\`) {
				// This is a UNC path, so we need to add 'UNC' to the path as well.
				path = Prefix + `UNC` + path[1:]
			} else {
				path = Prefix + path
			}
		}
	}
	return path
}
