package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

type TokenManager interface {
	GetToken() ([]byte, error)
	SetToken([]byte) error
}

type EncryptWrapper struct {
	aead     cipher.AEAD
	keyBytes []byte
}

func NewFromManager(tokenManager TokenManager, key []byte) (*EncryptWrapper, error) {
	encrypter, err := NewEncrypter(key)
	if err != nil {
		return nil, err
	}

	rootKey, err := tokenManager.GetToken()
	if err != nil {
		return nil, err
	}

	var rootKeyDecrypted []byte
	switch len(rootKey) {
	case 0:
		rootKeyDecrypted = make([]byte, 32)
		_, err := rand.Read(rootKeyDecrypted)
		if err != nil {
			return nil, err
		}

		rootKeyEnc, err := encrypter.Encrypt(rootKeyDecrypted)
		if err != nil {
			return nil, err
		}

		tokenManager.SetToken(rootKeyEnc)
	case 32 + aes.BlockSize + encrypter.aead.NonceSize():
		rootKeyDecrypted, err = encrypter.Decrypt(rootKey)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid key size")
	}

	encrypter, err = NewEncrypter(rootKeyDecrypted)
	if err != nil {
		return nil, err
	}

	return encrypter, nil
}

func NewEncrypter(key []byte) (*EncryptWrapper, error) {
	w := &EncryptWrapper{}
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return nil, err
	}

	w.keyBytes = key
	w.aead = aead
	return w, nil
}

func (w *EncryptWrapper) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := w.aead.Seal(nil, nonce, plaintext, nil)

	return append(nonce, ciphertext...), nil
}

func (w *EncryptWrapper) Decrypt(cipherText []byte) ([]byte, error) {
	nonce, ciphertext := cipherText[:12], cipherText[12:]

	plaintext, err := w.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
