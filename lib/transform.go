package icepacker

import (
	"bytes"
	"compress/gzip"

	"fmt"

	"io"
	"io/ioutil"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"

	"golang.org/x/crypto/pbkdf2"
)

// ChiperSettings records the settings of encryption/decryption
type CipherSettings struct {
	Key       string
	Salt      string
	Iteration int
}

// NewCipherSettings created a new CipherSettings instance with default values
func NewCipherSettings(key string) CipherSettings {
	return CipherSettings{
		Key:       key,
		Salt:      "icepacker",
		Iteration: 10000,
	}
}

// HashingKey is hashing the key with CipherSettings values
func HashingKey(settings CipherSettings) []byte {
	if settings.Iteration > 0 {
		k := pbkdf2.Key([]byte(settings.Key), []byte(settings.Salt), settings.Iteration, 16, sha1.New)
		return k
	}
	return []byte(settings.Key)
}

// decrypt is decrypting the content with the key
func decrypt(ciphertext []byte, key []byte) []byte {
	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// Before even testing the decryption,
	// if the text is too small, then it is incorrect
	if len(ciphertext) < aes.BlockSize {
		panic(fmt.Errorf("Text is too short: %d", len(ciphertext)))
	}

	// Get the 16 byte IV
	iv := ciphertext[:aes.BlockSize]

	// Remove the IV from the ciphertext
	ciphertext = ciphertext[aes.BlockSize:]

	// Return a decrypted stream
	stream := cipher.NewCFBDecrypter(block, iv)

	// Decrypt bytes from ciphertext
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext
}

// encrypt is encrypting the content with the key
func encrypt(plaintext []byte, key []byte) []byte {
	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// Empty array of 16 + plaintext length
	// Include the IV at the beginning
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	// Slice of first 16 bytes
	iv := ciphertext[:aes.BlockSize]

	// Write 16 rand bytes to fill iv
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	// Return an encrypted stream
	stream := cipher.NewCFBEncrypter(block, iv)

	// Encrypt bytes from plaintext to ciphertext
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext
}

// compress is compress the data with GZIP
func compress(data []byte) []byte {
	var b bytes.Buffer

	gz, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		panic(err)
	}

	if _, err := gz.Write(data); err != nil {
		panic(err)
	}

	if err := gz.Flush(); err != nil {
		panic(err)
	}

	if err := gz.Close(); err != nil {
		panic(err)
	}

	return b.Bytes()
}

// decompress is decompress the data with GUNZIP
func decompress(data []byte) []byte {
	b := bytes.NewReader(data)

	gz, err := gzip.NewReader(b)
	if err != nil {
		panic(err)
	}
	defer gz.Close()

	content, err := ioutil.ReadAll(gz)
	if err != nil {
		panic(err)
	}

	return []byte(content)
}

// TransformPack is transform the content of file to the package (encrypt, compress)
func TransformPack(data []byte, compression byte, encryption byte, key []byte) ([]byte, error) {
	var res []byte

	if len(data) == 0 {
		return data, nil
	}

	// Compression
	if compression == COMPRESS_GZIP {
		res = compress(data)
	} else {
		res = data
	}

	// Encryption
	if encryption == ENCRYPT_AES {
		res = encrypt(res, key)
	}

	return res, nil
}

// TransformUnpack is transform back the transformed file content to the real file
func TransformUnpack(data []byte, compression byte, encryption byte, key []byte) ([]byte, error) {
	var res []byte

	if len(data) == 0 {
		return data, nil
	}

	// Encryption
	if encryption == ENCRYPT_AES {
		res = decrypt(data, key)
	} else {
		res = data
	}

	// Compression
	if compression == COMPRESS_GZIP {
		res = decompress(res)
	}

	return res, nil
}
