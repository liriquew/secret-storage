package jwt

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/liriquew/secret_storage/server/internal/models"
)

var Secret string = "AnyEps"

// NewToken creates new JWT token for given user and app.
func NewToken(user *models.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["unm"] = user.Username

	tokenString, err := token.SignedString([]byte(Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func Validate(tokenString string) (string, error) {
	claims := &jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(Secret), nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	return (*claims)["unm"].(string), nil
}
