package service

import (
	"crypto/rand"
	"math/big"
	"strings"

	"github.com/gin-gonic/gin"
)

func extractPath(c *gin.Context) []string {
	username := c.Value(usernameKey).(string)
	queryPath := strings.Split(c.Query(pathParam), "/")

	path := make([]string, 0, len(queryPath)+1)
	path = append(path, username)

	for _, pathPart := range queryPath {
		if pathPart != "" {
			path = append(path, pathPart)
		}
	}

	return path
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
