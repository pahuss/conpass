package encoder

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

type Aes struct {
	key []byte
}

func (a *Aes) Encrypt(data []byte) ([]byte, error) {
	if len(a.key) == 0 {
		return nil, errors.New("empty encoder key")
	}
	cphr, err := aes.NewCipher(a.key)
	if err != nil {
		return data, err
	}
	gcm, err := cipher.NewGCM(cphr)
	if err != nil {
		return data, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return data, err
	}
	return gcm.Seal(nonce, nonce, data, nil), nil
}

func (a *Aes) Decrypt(data []byte) ([]byte, error) {
	if len(a.key) == 0 {
		return nil, errors.New("empty encoder key")
	}
	c, err := aes.NewCipher(a.key)
	if err != nil {
		return data, err
	}
	gcmDecrypt, err := cipher.NewGCM(c)
	if err != nil {
		return data, err
	}
	nonceSize := gcmDecrypt.NonceSize()
	if len(data) < nonceSize {
		return data, err
	}
	nonce, encryptedMessage := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcmDecrypt.Open(nil, nonce, encryptedMessage, nil)
	if err != nil {
		return data, err
	}
	return plaintext, nil
}
func (a *Aes) SetKey(key []byte) {
	a.key = key
}
