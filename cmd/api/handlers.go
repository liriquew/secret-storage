package main

import (
	"encoding/json"
	"errors"
	"kv-storage/pkg/encrypt_db"
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

type tokenJWT struct {
	Token string `json:"token,omitempty"`
}

const (
	keyUrlParam = "key"

	headerContentType = "Content-Type"
	jsonContentType   = "application/json"
)

func (api *API) getByKey(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("GET")

	key := r.Context().Value(keyUrlParam).(string)

	value, err := api.storage.Get([]byte(key))

	if err != nil {
		api.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if value == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("key not found"))
		return
	}

	w.Header().Set(headerContentType, jsonContentType)
	json.NewEncoder(w).Encode(kvData{key, string(value)})
}

func (api *API) setByKey(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("SET")

	var kv kvData
	err := json.NewDecoder(r.Body).Decode(&kv)

	if err != nil {
		api.errorLog.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if kv.Key == "" || kv.Value == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = api.storage.Set([]byte(kv.Key), []byte(kv.Value))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
}

func (api *API) deleteByKey(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("DELETE")

	key := r.Context().Value(keyUrlParam).(string)

	err := api.storage.Delete([]byte(key))
	if err != nil {
		api.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(200)
}

func (api *API) listKV(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("LIST")

	kvs, err := api.storage.List()

	if err != nil {
		api.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(kvs) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set(headerContentType, jsonContentType)
	json.NewEncoder(w).Encode(kvs)
}

func (api *API) signUp(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("Sign up")
	var user uData
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil || user.Username == "" || user.Password == "" {
		api.errorLog.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = api.storage.CreateNewUser(user.Username, user.Password)

	if errors.Is(err, encrypt_db.UserAlreadyExistErr) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(err.Error()))
		return
	}

	if err != nil {
		api.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var token string
	token, err = api.GenerateJWT(user.Username)

	if err != nil {
		api.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, jsonContentType)
	json.NewEncoder(w).Encode(tokenJWT{token})
}

func (api *API) signIn(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("Sign in")

	var user uData
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil || user.Username == "" || user.Password == "" {
		api.errorLog.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = api.storage.SelectUser(user.Username, user.Password)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	var token string
	token, err = api.GenerateJWT(user.Username)

	if err != nil {
		api.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Set(headerContentType, jsonContentType)
	json.NewEncoder(w).Encode(tokenJWT{token})
}

// далее обработчики для дебага

func (api *API) test(w http.ResponseWriter, r *http.Request) {
	api.storage.ListEncrypted()
}

func (api *API) showRootKey(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("Show root key")
	key, err := api.storage.GetRootToken()
	if err != nil {
		api.errorLog.Println(err)
		w.WriteHeader(500)
	}

	api.infoLog.Println("ROOTKEY:\t", key)
	w.WriteHeader(200)
}
