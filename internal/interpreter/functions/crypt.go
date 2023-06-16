package functions

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
)

func Sha1(args [][]byte) (ret []byte, err error) {
	if len(args) != 1 {
		err = errors.New("exactly 1 arg expected")
		return
	}

	ret, err = EncodeHashToHex(sha1.New(), string(args[0]))
	return
}

func Sha256(args [][]byte) (ret []byte, err error) {
	if len(args) != 1 {
		err = errors.New("exactly 1 arg expected")
		return
	}

	ret, err = EncodeHashToHex(sha256.New(), string(args[0]))
	if err != nil {
		return
	}
	return
}

func Sha512(args [][]byte) (ret []byte, err error) {
	if len(args) != 1 {
		err = errors.New("exactly 1 arg expected")
		return
	}

	ret, err = EncodeHashToHex(sha512.New(), string(args[0]))
	return
}

func Md5(args [][]byte) (ret []byte, err error) {
	if len(args) != 1 {
		err = errors.New("exactly 1 arg expected")
		return
	}

	ret, err = EncodeHashToHex(md5.New(), string(args[0]))
	return
}
