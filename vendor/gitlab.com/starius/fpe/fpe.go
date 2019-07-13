package fpe

import (
	"encoding/binary"
	"math/bits"
)

// A BlockCipher represents an implementation of block cipher using a given key.
type BlockCipher interface {
	// BlockSize returns the cipher's block size.
	BlockSize() int

	// Encrypt encrypts the first block in src into dst.
	// Dst and src must overlap entirely or not at all.
	Encrypt(dst, src []byte)
}

// FeistelRounds is the number of rounds in random Feistel scheme.
// See Jacques Patarin CRYPTO'03 for the number.
const FeistelRounds = 10

// SimpleEncrypt encrypts the plaintext number into another number in [0;maxvalue].
func SimpleEncrypt(blockCipher BlockCipher, plaintext, maxvalue uint64) uint64 {
	return Encrypt(blockCipher, plaintext, maxvalue, 0, FeistelRounds)
}

// SimpleDecrypt decrypts the ciphertext number back.
func SimpleDecrypt(blockCipher BlockCipher, ciphertext, maxvalue uint64) uint64 {
	return Decrypt(blockCipher, ciphertext, maxvalue, 0, FeistelRounds)
}

// Encrypt encrypts the plaintext number into another number in [0;maxvalue].
// You can specify the number of feistel network rounds and a tweak (publicly
// known modifier of the algorithm).
func Encrypt(blockCipher BlockCipher, plaintext, maxvalue, tweak uint64, feistelRounds int) uint64 {
	oddRounds := (feistelRounds % 2) == 1
	return encryptDecrypt(blockCipher, plaintext, maxvalue, tweak, true, oddRounds, 0, feistelRounds, 1)
}

// Decrypt decrypts the ciphertext number back.
func Decrypt(blockCipher BlockCipher, ciphertext, maxvalue, tweak uint64, feistelRounds int) uint64 {
	oddRounds := (feistelRounds % 2) == 1
	return encryptDecrypt(blockCipher, ciphertext, maxvalue, tweak, false, oddRounds, feistelRounds-1, -1, -1)
}

var zeros = make([]byte, 64)

func encryptDecrypt(blockCipher BlockCipher, value, maxvalue, tweak uint64, encrypt, oddRounds bool, startRound, endRound, delta int) uint64 {
	if blockCipher.BlockSize() < 16 && tweak != 0 {
		panic("tweak is not supported for block ciphers with blocks smaller than 16")
	}
	totalBits := uint(bits.Len64(maxvalue))
	if totalBits < 2 {
		panic("domain is too small")
	}
	for {
		value = feistelNetwork(blockCipher, value, tweak, totalBits, encrypt, oddRounds, startRound, endRound, delta)
		if value <= maxvalue {
			return value
		}
	}
}

func feistelNetwork(blockCipher BlockCipher, value, tweak uint64, totalBits uint, encrypt, oddRounds bool, startRound, endRound, delta int) uint64 {
	buffer := make([]byte, 64)[:blockCipher.BlockSize()]
	leftBits := totalBits / 2
	rightBits := totalBits - leftBits
	if !encrypt && oddRounds {
		leftBits, rightBits = rightBits, leftBits
	}
	left := uint32(value >> rightBits)
	right := uint32(value & ((1 << rightBits) - 1))
	if !encrypt {
		leftBits, rightBits = rightBits, leftBits
		left, right = right, left
	}
	// Perform random Feistel scheme.
	for round := startRound; round != endRound; round += delta {
		// Encrypt old R, XOR the result with old L and assign to new R. Assign new L to old R.
		copy(buffer, zeros)
		binary.LittleEndian.PutUint32(buffer[0:4], right)
		buffer[5] = byte(round) // To make all round functions different.
		if blockCipher.BlockSize() >= 16 {
			binary.LittleEndian.PutUint64(buffer[8:16], tweak)
		}
		blockCipher.Encrypt(buffer, buffer)
		cipherText := binary.LittleEndian.Uint32(buffer[:4])
		xor := cipherText ^ left
		leftBits, rightBits = rightBits, leftBits
		left = right
		right = xor & ((1 << rightBits) - 1)
	}
	if !encrypt {
		rightBits = leftBits // leftBits = rightBits
		left, right = right, left
	}
	return (uint64(left) << rightBits) | uint64(right)
}
