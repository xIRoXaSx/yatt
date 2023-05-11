package functions

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
)

func Sha1(fn string, args [][]byte) (ret []byte, err error) {
	if len(args) != 1 {
		err = fmt.Errorf("%s: exactly 1 arg expected", fn)
		return
	}

	ret, err = EncodeHashToHex(sha1.New(), string(args[0]))
	return
}

func Sha256(fn string, args [][]byte) (ret []byte, err error) {
	if len(args) != 1 {
		err = fmt.Errorf("%s: exactly 1 arg expected", fn)
		return
	}

	ret, err = EncodeHashToHex(sha256.New(), string(args[0]))
	if err != nil {
		return
	}
	return
}

func Sha512(fn string, args [][]byte) (ret []byte, err error) {
	if len(args) != 1 {
		err = fmt.Errorf("%s: exactly 1 arg expected", fn)
		return
	}

	ret, err = EncodeHashToHex(sha512.New(), string(args[0]))
	return
}

func Md5(fn string, args [][]byte) (ret []byte, err error) {
	if len(args) != 1 {
		err = fmt.Errorf("%s: exactly 1 arg expected", fn)
		return
	}

	ret, err = EncodeHashToHex(md5.New(), string(args[0]))
	return
}
