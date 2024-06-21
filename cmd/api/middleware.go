package main

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"net/http"
	"strings"
	"time"
)

func (api *API) AuthRequired(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, err := CheckJWT(r)

		if err != nil {
			api.errorLog.Println("error checking JWT token:", err)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		// проверка, существует ли такой пользователь
		isExist, err := api.storage.CheckUser(username)

		if err != nil {
			api.errorLog.Println("Check user error:", err)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		if !isExist {
			api.errorLog.Println("user not found")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func CheckJWT(r *http.Request) (string, error) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		return "", errors.New("Missing Authorization header")
	}

	tokenString, found := strings.CutPrefix(tokenString, "Bearer ")

	if !found {
		return "", errors.New("Missing Authorization header")
	}

	claims := &jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JWTsecretKey, nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("Invalid token")
	}

	username := (*claims)["sub"].(string)
	return username, nil
}

func (api *API) GenerateJWT(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 12).Unix(),
		"iat": time.Now().Unix(),
		"sub": username,
	}

	tokenString, err := token.SignedString(JWTsecretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
