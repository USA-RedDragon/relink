package relink

import (
	"io"
	"os"

	"golang.org/x/crypto/blake2b"
)

func HashFile(filePath string) (ret []byte, err error) {
	b2b, err := blake2b.New512(nil)
	if err != nil {
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer f.Close()

	for {
		buf := make([]byte, 4096)
		readN, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if readN == 0 {
			break
		}
		writeN, err := b2b.Write(buf[:readN])
		if err != nil {
			return nil, err
		}
		if writeN != readN {
			return nil, io.ErrShortBuffer
		}
	}

	return b2b.Sum(nil), nil
}
