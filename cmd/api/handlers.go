package main

import (
	"encoding/json"
	"net/http"
)

type kvData struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type uData struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

const (
	headerContentType = "Content-Type"
	jsonContentType   = "application/json"
	jwtAuthToken      = "token"
)

func (app *API) getByKey(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Println("GET by key")

	key := r.URL.Query().Get("key")
	if key == "" {
		w.WriteHeader(400)
		return
	}

	value, err := app.storage.Get([]byte(key))
	if err != nil {
		w.WriteHeader(500)
		return
	}

	if value == nil {
		w.WriteHeader(404)
		return
	}

	w.Header().Set(headerContentType, jsonContentType)
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(kvData{"", string(value)})
}

func (app *API) setValueByKey(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Println("SET by key")

	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")

	if key == "" {
		w.WriteHeader(400)
		return
	}

	err := app.storage.Put([]byte(key), []byte(value))
	if err != nil {
		w.WriteHeader(400)
		return
	}

	w.WriteHeader(200)
}

func (app *API) deleteByKey(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Println("delete by key")

	key := r.URL.Query().Get("key")
	if key == "" {
		w.WriteHeader(400)
		return
	}

	err := app.storage.Delete([]byte(key))
	if err != nil {
		w.WriteHeader(500)
	}
	w.WriteHeader(200)
}

func (app *API) test(w http.ResponseWriter, r *http.Request) {
	app.storage.List()
}

func (app *API) signUp(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Println("Sign up")
	var user uData
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		w.WriteHeader(500)
		return
	}

	if user.Username == "" || user.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = app.storage.CreateNewUser(user.Username, user.Password)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte(err.Error()))
		return
	}

	var token string
	token, err = app.GenerateJWT(user.Username)

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}

	w.Header().Set(jwtAuthToken, token)
	w.WriteHeader(200)
}

func (app *API) signIn(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Println("Sign in")

	var user uData
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		w.WriteHeader(500)
		return
	}
	if user.Username == "" || user.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = app.storage.SelectUser(user.Username, user.Password)

	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte(err.Error()))
		return
	}

	var token string
	token, err = app.GenerateJWT(user.Username)

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}

	w.Header().Set(jwtAuthToken, token)
	w.WriteHeader(200)
}
