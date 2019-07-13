package encrypt

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"gitlab.com/NebulousLabs/fastrand"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/crypto/hkdf"
)

var salt = []byte("salt for gitlab.com/starius/encrypt-autocert-cache")

const keyPrefixLen = 32

type EncryptedCache struct {
	keyPrefix []byte
	valueAEAD cipher.AEAD
	backend   autocert.Cache
}

func NewEncryptedCache(backendCache autocert.Cache, key []byte) (*EncryptedCache, error) {
	// Key derivation.
	keyHKDF := hkdf.New(sha256.New, key, salt, []byte("key"))
	keyPrefix := make([]byte, keyPrefixLen)
	if _, err := io.ReadFull(keyHKDF, keyPrefix); err != nil {
		return nil, err
	}
	valueHKDF := hkdf.New(sha256.New, key, salt, []byte("value"))
	valueAESKey := make([]byte, 32)
	if _, err := io.ReadFull(valueHKDF, valueAESKey); err != nil {
		return nil, err
	}

	// AEAD for value.
	valueAES, err := aes.NewCipher(valueAESKey)
	if err != nil {
		return nil, err
	}
	valueAEAD, err := cipher.NewGCM(valueAES)
	if err != nil {
		return nil, err
	}

	return &EncryptedCache{
		keyPrefix: keyPrefix,
		valueAEAD: valueAEAD,
		backend:   backendCache,
	}, nil
}

func (c *EncryptedCache) hashKey(key string) string {
	keyHash := sha256.Sum256([]byte(key))
	var buffer [keyPrefixLen + sha256.Size]byte
	copy(buffer[:keyPrefixLen], c.keyPrefix)
	copy(buffer[keyPrefixLen:], keyHash[:])
	hash2 := sha256.Sum256(buffer[:])
	return base64.RawURLEncoding.EncodeToString(hash2[:])
}

func (c *EncryptedCache) Get(ctx context.Context, key string) ([]byte, error) {
	keyHash := c.hashKey(key)
	encryptedValue, err := c.backend.Get(ctx, keyHash)
	if err != nil {
		return nil, err
	}
	nonceSize := c.valueAEAD.NonceSize()
	if len(encryptedValue) < nonceSize {
		return nil, fmt.Errorf("the value is too short")
	}
	nonce := encryptedValue[:nonceSize]
	return c.valueAEAD.Open(nil, nonce, encryptedValue[nonceSize:], []byte(keyHash))
}

func (c *EncryptedCache) Put(ctx context.Context, key string, data []byte) error {
	keyHash := c.hashKey(key)

	nonceSize := c.valueAEAD.NonceSize()
	nonce := fastrand.Bytes(nonceSize)
	// Append the ciphertext to nonce.
	encryptedValue := c.valueAEAD.Seal(nonce, nonce, data, []byte(keyHash))

	return c.backend.Put(ctx, keyHash, encryptedValue)
}

func (c *EncryptedCache) Delete(ctx context.Context, key string) error {
	keyHash := c.hashKey(key)
	return c.backend.Delete(ctx, keyHash)
}
