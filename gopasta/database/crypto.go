package database

import (
	"crypto/cipher"
	"encoding/binary"
	"fmt"
)

type EncryptedFile struct {
	block cipher.Block
	impl  File
}

func WrapInCTR(block cipher.Block, impl File) *EncryptedFile {
	return &EncryptedFile{block, impl}
}

func applyCTR(block cipher.Block, p []byte, off int64) {
	tailLen := off % int64(block.BlockSize())
	off -= tailLen
	off /= int64(block.BlockSize())
	iv := make([]byte, block.BlockSize())
	binary.BigEndian.PutUint64(iv[len(iv)-8:], uint64(off))
	stream := cipher.NewCTR(block, iv)
	tail := make([]byte, tailLen)
	stream.XORKeyStream(tail, tail)
	stream.XORKeyStream(p, p)
}

func allZeros(p []byte) bool {
	for _, c := range p {
		if c != 0 {
			return false
		}
	}
	return true
}

func (f *EncryptedFile) ReadAt(p []byte, off int64) (n int, err error) {
	n, err = f.impl.ReadAt(p, off)
	if err != nil {
		return n, err
	}
	if allZeros(p) {
		return n, fmt.Errorf("data was wiped")
	}
	applyCTR(f.block, p, off)
	return n, nil
}

func (f *EncryptedFile) WriteAt(p []byte, off int64) (n int, err error) {
	applyCTR(f.block, p, off)
	return f.impl.WriteAt(p, off)
}
