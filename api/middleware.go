package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type usernameInterface struct{}

var JWTsecretKey []byte

func (api *API) AuthRequired(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, err := CheckJWT(r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if username == "root" {
			ctx := context.WithValue(r.Context(), usernameInterface{}, "")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// проверка, существует ли такой пользователь
		isExist, err := api.storage.CheckUser(username)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if !isExist {
			http.Error(w, "user not exist", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), usernameInterface{}, username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (api *API) RootRequired(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, err := CheckJWT(r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if username != "root" {
			http.Error(w, "root required", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), usernameInterface{}, username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (api *API) shamirRequired(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if api.storage == nil {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("not enough key parts"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func CheckJWT(r *http.Request) (string, error) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	tokenString, found := strings.CutPrefix(tokenString, "Bearer ")

	if !found {
		return "", fmt.Errorf("missing Authorization header")
	}

	claims := &jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JWTsecretKey, nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
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
