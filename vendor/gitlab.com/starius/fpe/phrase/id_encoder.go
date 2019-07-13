package phrase

import (
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"
)

// IDEncoder converts between an integer ID and words or encrypted form.
type IDEncoder struct {
	encrypt, decrypt PRP
	base             uint64
	dict             *Dict

	longCipher cipher.Block
}

// NewIDEncoder creates new IDEncoder. It accepts separate block ciphers to
// encrypt into words and to encrypt into long string and also the list of words.
func NewIDEncoder(wordsCipher, longCipher cipher.Block, words []string) (*IDEncoder, error) {
	feistelRounds := 4
	encrypt, decrypt := MakePRP(wordsCipher, feistelRounds)
	dict, err := MakeDict(words)
	if err != nil {
		return nil, err
	}
	return &IDEncoder{
		encrypt:    encrypt,
		decrypt:    decrypt,
		base:       uint64(len(words)),
		dict:       dict,
		longCipher: longCipher,
	}, nil
}

// Encode encodes ID into words or long string depending on longID flag.
func (e *IDEncoder) Encode(id uint64, longID bool) (text string, err error) {
	if longID {
		buffer := make([]byte, e.longCipher.BlockSize())
		binary.LittleEndian.PutUint64(buffer[0:8], id)
		e.longCipher.Encrypt(buffer, buffer)
		return base64.RawURLEncoding.EncodeToString(buffer), nil
	}
	numbers, err := ToNumbersList(id, e.base, e.encrypt)
	if err != nil {
		return "", err
	}
	words, err := e.dict.ToWords(numbers)
	if err != nil {
		return "", err
	}
	return strings.Join(words, "-"), nil
}

// Decode accepts a string returned by Encode and returns original ID and longID flag.
func (e *IDEncoder) Decode(text string) (id uint64, longID bool, err error) {
	words := strings.Split(text, "-")
	numbers, err := e.dict.ToNumbers(words)
	if err == nil {
		id, err = FromNumbersList(numbers, e.base, e.decrypt)
		return id, false, err
	}
	// Maybe it is a long id?
	buffer, err := base64.RawURLEncoding.DecodeString(text)
	if err != nil {
		return 0, false, err
	}
	if len(buffer) != e.longCipher.BlockSize() {
		return 0, false, fmt.Errorf("bad ID length")
	}
	e.longCipher.Decrypt(buffer, buffer)
	for i := 8; i < len(buffer); i++ {
		if buffer[i] != 0 {
			return 0, false, fmt.Errorf("invalid ID")
		}
	}
	id = binary.LittleEndian.Uint64(buffer[0:8])
	return id, true, nil
}
