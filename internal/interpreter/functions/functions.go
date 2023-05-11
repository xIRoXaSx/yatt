package functions

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"hash"
	"os"
	"strconv"
)

func ParseFloats(args [][]byte) (floats []float64, err error) {
	floats = make([]float64, len(args))
	for i := range args {
		floats[i], err = strconv.ParseFloat(string(bytes.TrimSpace(args[i])), 64)
		if err != nil {
			err = fmt.Errorf("%v: args=%s", err, bytes.Join(args, []byte(", ")))
			return
		}
	}
	return
}

func EncodeHashToHex(h hash.Hash, file string) (sum []byte, err error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return
	}

	h.Write(b)
	s := h.Sum(nil)
	sum = make([]byte, hex.EncodedLen(len(s)))
	hex.Encode(sum, s)
	return
}

// TrimQuotes trims spaces, prefix and suffix separately to allow symbol escaping.
func TrimQuotes(val []byte) (ret []byte) {
	sq := []byte("'")
	dq := []byte("\"")
	ret = bytes.TrimSpace(val)
	ret = bytes.TrimPrefix(ret, dq)
	ret = bytes.TrimPrefix(ret, sq)
	ret = bytes.TrimSuffix(ret, dq)
	ret = bytes.TrimSuffix(ret, sq)
	return
}
