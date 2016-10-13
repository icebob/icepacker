package icepacker

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTransformShaKey(t *testing.T) {

	Convey("Should hashing the key", t, func() {

		So(HashingKey(CipherSettings{Key: "password"}), ShouldResemble, []byte("password"))
		So(HashingKey(CipherSettings{Key: "password"}), ShouldResemble, []byte("password"))
		So(HashingKey(CipherSettings{Key: "password", Iteration: 100}), ShouldResemble, []uint8{146, 185, 30, 84, 124, 150, 121, 57, 156, 174, 62, 87, 121, 108, 63, 7})
		So(HashingKey(CipherSettings{Key: "password", Iteration: 1000}), ShouldResemble, []uint8{151, 35, 89, 28, 209, 207, 236, 176, 239, 215, 176, 39, 234, 147, 177, 0})
		So(HashingKey(CipherSettings{Key: "password", Iteration: 1000, Salt: "salt"}), ShouldResemble, []uint8{110, 136, 190, 139, 173, 126, 174, 157, 158, 16, 170, 6, 18, 36, 3, 79})

		So(HashingKey(NewCipherSettings("password")), ShouldResemble, []uint8{30, 105, 154, 144, 241, 102, 33, 180, 53, 178, 108, 142, 123, 221, 221, 61})
		So(HashingKey(NewCipherSettings("123123123")), ShouldResemble, []uint8{182, 51, 251, 197, 217, 163, 32, 193, 27, 21, 18, 30, 85, 41, 161, 15})
	})

}

func TestTransformEncryptDecrypt(t *testing.T) {
	key := HashingKey(CipherSettings{Key: "password", Iteration: 500})
	origText := "Original plain text"

	Convey("Should equal the encrypted & decrypted text", t, func() {
		cipher := encrypt([]byte(origText), key)
		So(string(cipher), ShouldNotEqual, origText)
		text := decrypt(cipher, key)
		So(string(text), ShouldEqual, origText)
	})
}

func TestTransformCompressDecompress(t *testing.T) {
	origText := "Original plain data"

	Convey("Should equal the compressed & decompressed text", t, func() {
		compressed := compress([]byte(origText))
		So(string(compressed), ShouldNotEqual, origText)
		text := decompress(compressed)
		So(string(text), ShouldEqual, origText)
	})
}

func TestTransformPack(t *testing.T) {
	key := HashingKey(CipherSettings{Key: "password2", Iteration: 500})
	origText := "Original plain data"

	Convey("Should equal the ENCRYPT_NONE and COMPRESS_NONE text", t, func() {
		transformed, err := TransformPack([]byte(origText), COMPRESS_NONE, ENCRYPT_NONE, key)
		So(err, ShouldBeNil)
		So(string(transformed), ShouldEqual, origText)

		detransformed, err := TransformUnpack(transformed, COMPRESS_NONE, ENCRYPT_NONE, key)
		So(err, ShouldBeNil)
		So(string(detransformed), ShouldEqual, origText)
	})

	Convey("Should equal the encrypted and compressed text", t, func() {

		Convey("test every encryption and compression combination", func() {

			var tests = []struct {
				compress byte
				encrypt  byte
			}{
				{COMPRESS_NONE, ENCRYPT_AES},
				{COMPRESS_GZIP, ENCRYPT_NONE},
				{COMPRESS_GZIP, ENCRYPT_AES},
			}

			for _, test := range tests {
				transformed, err := TransformPack([]byte(origText), test.compress, test.encrypt, key)
				So(err, ShouldBeNil)
				So(string(transformed), ShouldNotEqual, origText)

				detransformed, err := TransformUnpack(transformed, test.compress, test.encrypt, key)
				So(err, ShouldBeNil)
				So(string(detransformed), ShouldEqual, origText)
			}

		})

	})

}
