package functions

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"os"
)

// SHA1 creates a SHA1 sum of the given file, provided by arg at index 0.
func SHA1(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthExact(args, 1)
	if err != nil {
		return
	}

	return encodeHashToHex(sha1.New(), string(args[0]))
}

// SHA256 creates a SHA256 sum of the given file, provided by arg at index 0.
func SHA256(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthExact(args, 1)
	if err != nil {
		return
	}

	return encodeHashToHex(sha256.New(), string(args[0]))
}

// SHA512 creates a SHA512 sum of the given file, provided by arg at index 0.
func SHA512(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthExact(args, 1)
	if err != nil {
		return
	}

	return encodeHashToHex(sha512.New(), string(args[0]))

}

// MD5 creates a MD5 sum of the given file, provided by arg at index 0.
func MD5(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthExact(args, 1)
	if err != nil {
		return
	}

	return encodeHashToHex(md5.New(), string(args[0]))
}

//
// Helper.
//

func encodeHashToHex(h hash.Hash, file string) (sum []byte, err error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return
	}

	_, err = h.Write(b)
	if err != nil {
		return
	}
	s := h.Sum(nil)
	sum = make([]byte, hex.EncodedLen(len(s)))
	_ = hex.Encode(sum, s)
	return
}
