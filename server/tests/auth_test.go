package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/liriquew/secret_storage/server/internal/lib/jwt"
	"github.com/liriquew/secret_storage/server/internal/models"
	"github.com/liriquew/secret_storage/server/tests/suite"
	"github.com/stretchr/testify/assert"
)

const (
	contentType     = "Content-Type"
	applicationJSON = "application/json"

	StatusUnauthorized = "401 Unauthorized"
)

type JWT struct {
	Token string `json:"jwt-token"`
}

func TestSignUp(t *testing.T) {
	ts := suite.New(t)

	user := models.User{
		Username: gofakeit.Username(),
		Password: gofakeit.JobTitle(),
	}

	buf, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/signup", ts.GetURL()), bytes.NewBuffer(buf))
	req.Header.Set(contentType, applicationJSON)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, StatusOK, resp.Status)

	var token JWT

	err = json.NewDecoder(resp.Body).Decode(&token)
	assert.NoError(t, err)

	usernameFromJWT, err := jwt.Validate(token.Token)
	assert.NoError(t, err)

	assert.Equal(t, user.Username, usernameFromJWT)
}

func TestSignIn(t *testing.T) {
	ts := suite.New(t)

	user := models.User{
		Username: gofakeit.Username(),
		Password: gofakeit.JobTitle(),
	}

	buf, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/signup", ts.GetURL()), bytes.NewBuffer(buf))
	req.Header.Set(contentType, applicationJSON)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, StatusOK, resp.Status)

	t.Run("Success", func(t *testing.T) {
		req, _ = http.NewRequest("POST", fmt.Sprintf("%s/signin", ts.GetURL()), bytes.NewBuffer(buf))
		req.Header.Set(contentType, applicationJSON)

		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, StatusOK, resp.Status)

		var token JWT

		err = json.NewDecoder(resp.Body).Decode(&token)
		assert.NoError(t, err)

		usernameFromJWT, err := jwt.Validate(token.Token)
		assert.NoError(t, err)

		assert.Equal(t, user.Username, usernameFromJWT)
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		user1 := models.User{
			Username: user.Username,
			Password: gofakeit.FarmAnimal(),
		}

		buf, _ := json.Marshal(user1)

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/signin", ts.GetURL()), bytes.NewBuffer(buf))
		req.Header.Set(contentType, applicationJSON)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, StatusUnauthorized, resp.Status)

	})
}

type UserWithToken struct {
	User  models.User
	Token string
}

func CreateUser(t *testing.T, ts *suite.Suite) *UserWithToken {
	user := models.User{
		Username: gofakeit.Username(),
		Password: gofakeit.JobTitle(),
	}

	buf, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/signup", ts.GetURL()), bytes.NewBuffer(buf))
	req.Header.Set(contentType, applicationJSON)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, StatusOK, resp.Status)
	var token JWT

	err = json.NewDecoder(resp.Body).Decode(&token)
	assert.NoError(t, err)

	return &UserWithToken{
		User:  user,
		Token: token.Token,
	}
}
