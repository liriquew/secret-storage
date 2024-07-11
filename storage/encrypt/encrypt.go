package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"secret-storage/storage/bolt"
	"secret-storage/storage/shamir"
)

type Wrapper struct {
	aead     cipher.AEAD
	keyBytes []byte
}

func NewWrapper(db *bolt.Bolt, secretParts [][]byte) (*Wrapper, error) {
	secretKey, err := shamir.Combine(secretParts)
	if err != nil {
		return nil, err
	}
	shamirWrapper, err := getAesGcmByKeyBytes(secretKey)
	if err != nil {
		return nil, err
	}

	rootKey, err := db.GetToken()
	if err != nil {
		return nil, err
	}

	var rootKeyDec []byte
	switch len(rootKey) {
	case 0:
		rootKeyDec = make([]byte, 32)
		_, err := rand.Read(rootKeyDec)
		if err != nil {
			return nil, err
		}

		rootKeyEnc, err := shamirWrapper.Encrypt(rootKeyDec)
		if err != nil {
			return nil, err
		}

		db.SetToken(rootKeyEnc)
	case 32 + aes.BlockSize + shamirWrapper.aead.NonceSize():
		rootKeyDec, err = shamirWrapper.Decrypt(rootKey)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid key size")
	}

	wrapper, err := getAesGcmByKeyBytes(rootKeyDec)
	if err != nil {
		return nil, err
	}

	return wrapper, nil
}

func getAesGcmByKeyBytes(key []byte) (*Wrapper, error) {
	w := &Wrapper{}
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

func (w *Wrapper) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := w.aead.Seal(nil, nonce, plaintext, nil)

	return append(nonce, ciphertext...), nil
}

func (w *Wrapper) Decrypt(cipherText []byte) ([]byte, error) {
	nonce, ciphertext := cipherText[:12], cipherText[12:]

	plaintext, err := w.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
