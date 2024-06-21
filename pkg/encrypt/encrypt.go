package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"kv-storage/pkg/bolt"
)

type Wrapper struct {
	aead     cipher.AEAD
	keyBytes []byte
}

type RecordInfo struct {
	Ciphertext []byte
}

func NewWrapper(path string) (*Wrapper, error) {
	db, err := bolt.New(path)
	if err != nil {
		return nil, err
	}

	token, err := db.GetToken()
	if err != nil {
		return nil, err
	}

	var rootKey []byte

	switch len(token) {
	case 0:
		newKey := make([]byte, 32)
		_, err := rand.Read(newKey)
		if err != nil {
			return nil, err
		}
		rootKey = newKey
		db.SetToken(newKey)
	case 32:
		rootKey = token[:32]
	default:
		return nil, errors.New("Invalid key size")
	}

	if err = db.Close(); err != nil {
		return nil, errors.New("failed to close db")
	}

	wrapper, err := GetAesGcmByKeyBytes(rootKey)

	if err != nil {
		return nil, err
	}
	return wrapper, nil
}

func GetAesGcmByKeyBytes(key []byte) (*Wrapper, error) {
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

	fmt.Println(nonce)
	fmt.Println(ciphertext)

	return append(nonce, ciphertext...), nil
}

func (w *Wrapper) Decrypt(cipherText []byte) ([]byte, error) {
	nonce, ciphertext := cipherText[:12], cipherText[12:]

	fmt.Println(nonce)
	fmt.Println(ciphertext)

	plaintext, err := w.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
