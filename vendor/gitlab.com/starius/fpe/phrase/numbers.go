package phrase

import (
	"errors"

	"gitlab.com/starius/fpe"
)

// ErrOverflow is returned by ToNumbersList and FromNumbersList if overflow happens.
var ErrOverflow = errors.New("overflow")

// PRP is pseudorandom permutation (decrypt is inverse permutation of encrypt).
type PRP func(input, maxvalue uint64) uint64

// ToNumbersList encodes the number as list of numbers from range [0;base).
// First base numbers are encoded as a list with one element which is the
// result of encrypt PRP. Next base^2 numbers are encoded as 2-element list
// representing encrypt(number-base) in positional numeral system of base base. Etc.
func ToNumbersList(number, base uint64, encrypt PRP) (list []uint64, err error) {
	domainSize := uint64(1)
	digits := 0
	for {
		domainSize1 := domainSize * base
		if domainSize1/base != domainSize {
			return nil, ErrOverflow
		}
		domainSize = domainSize1
		digits++
		if number < domainSize {
			number = encrypt(number, domainSize-1)
			return toList(number, base, digits), nil
		}
		number -= domainSize
	}
}

func toList(number, base uint64, digits int) (list []uint64) {
	list = make([]uint64, digits)
	for i := 0; i < digits; i++ {
		list[digits-i-1] = number % base
		number /= base
	}
	return
}

// FromNumbersList converts the list of numbers from range [0;base) to the number.
// It is inverse of ToNumbersList.
func FromNumbersList(list []uint64, base uint64, decrypt PRP) (number uint64, err error) {
	number, err = fromList(list, base)
	if err != nil {
		return 0, err
	}
	domainSize := uint64(1)
	sum := uint64(0)
	for range list {
		domainSize1 := domainSize * base
		if domainSize1/base != domainSize {
			return 0, ErrOverflow
		}
		domainSize = domainSize1
		sum += domainSize
	}
	sum -= domainSize
	number = decrypt(number, domainSize-1)
	number += sum
	return number, nil
}

func fromList(list []uint64, base uint64) (number uint64, err error) {
	for _, digit := range list {
		if digit >= base {
			return 0, ErrOverflow
		}
		number1 := number * base
		if number1/base != number {
			return 0, ErrOverflow
		}
		number2 := number1 + digit
		if number2 < number1 {
			return 0, ErrOverflow
		}
		number = number2
	}
	return number, nil
}

// MakePRP returns encrypt and decrypt PRPs for ToNumbersList and FromNumbersList.
// It accepts block cipher as defined by fpe package. See fpr/README.md for
// values of feistelRounds that make sense.
func MakePRP(blockCipher fpe.BlockCipher, feistelRounds int) (encrypt, decrypt PRP) {
	encrypt = func(input, maxvalue uint64) uint64 {
		return fpe.Encrypt(blockCipher, input, maxvalue, maxvalue, feistelRounds)
	}
	decrypt = func(input, maxvalue uint64) uint64 {
		return fpe.Decrypt(blockCipher, input, maxvalue, maxvalue, feistelRounds)
	}
	return
}
