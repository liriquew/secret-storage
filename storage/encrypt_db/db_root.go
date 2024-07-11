package encrypt_db

import (
	"crypto/rand"
	"math/big"
)

type DBinfo struct {
	Storage  *BoltEncrypt
	RootPass string
}

func GeneratePassword(length int) ([]byte, error) {
	const (
		lowerBytes = "abcdefghijklmnopqrstuvwxyz"
		upperBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digitBytes = "0123456789"
	)
	availableBytes := lowerBytes + upperBytes + digitBytes
	password := make([]byte, length)
	byteLength := byte(len(availableBytes))

	for i := range password {
		randomByte, err := rand.Int(rand.Reader, big.NewInt(int64(byteLength)))
		if err != nil {
			return nil, err
		}
		password[i] = availableBytes[randomByte.Int64()]
	}

	return password, nil
}
