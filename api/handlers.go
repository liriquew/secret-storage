package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"secret-storage/storage/encrypt_db"
	"strings"
)

type kvData struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type uData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type tokenJWT struct {
	Token string `json:"token"`
}

const (
	headerContentType = "Content-Type"
	jsonContentType   = "application/json"
	dbPath            = "./data.db"
	JWTname           = "jwtname"
)

var (
	kvBucketName   = []byte("kv")
	metaBucketName = []byte("meta")
)

func (api *API) setByKey(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("SET")

	username := r.Context().Value(usernameInterface{}).(string)
	prefix := r.Context().Value(pathPartsInterface{}).([]string)

	api.infoLog.Println(username, prefix)

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

	api.infoLog.Println(kv)
	err = api.storage.Set(username, prefix, kv.Key, kv.Value, kvBucketName)
	if err != nil {
		api.errorLog.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	api.infoLog.Println("SET OK")
}

func (api *API) getByKey(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("GET")
	username := r.Context().Value(usernameInterface{}).(string)
	pathParts := r.Context().Value(pathPartsInterface{}).([]string)

	if len(pathParts) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	key := pathParts[len(pathParts)-1]
	prefix := pathParts[:len(pathParts)-1]

	api.infoLog.Println(username, prefix, key)

	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
	}

	value, err := api.storage.Get(username, prefix, key, kvBucketName)

	if err != nil {
		api.errorLog.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set(headerContentType, jsonContentType)
	json.NewEncoder(w).Encode(kvData{key, string(value)})
	api.infoLog.Println("GET OK")
}

func (api *API) deleteByKey(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("DELETE")

	username := r.Context().Value(usernameInterface{}).(string)
	pathParts := r.Context().Value(pathPartsInterface{}).([]string)

	key := pathParts[len(pathParts)-1]
	prefix := pathParts[:len(pathParts)-1]

	api.infoLog.Println(username, prefix, key)

	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
	}

	deletedPartsCount, err := api.storage.Delete(username, prefix, key, kvBucketName)
	if err != nil {
		api.errorLog.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, _ := json.Marshal(deletedPartsCount)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	api.infoLog.Println("DEL OK")
}

func (api *API) listKV(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("LIST")

	username := r.Context().Value(usernameInterface{}).(string)
	path, _ := strings.CutPrefix(r.URL.Path, "/api/list/")
	prefix := strings.Split(path, "/")

	api.infoLog.Println(username, len(prefix), prefix)

	if len(prefix) != 0 && prefix[len(prefix)-1] == "" {
		prefix = prefix[:len(prefix)-1]
	}

	BucketInfo, err := api.storage.List(username, prefix)

	if err != nil {
		api.errorLog.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if len(BucketInfo.Buckets) == 0 && len(BucketInfo.KVs) == 0 {
		http.Error(w, "Empty bucket", http.StatusNotFound)
		return
	}

	w.Header().Set(headerContentType, jsonContentType)
	json.NewEncoder(w).Encode(BucketInfo)
}

func (api *API) listRecursion(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("LIST -r")

	username := r.Context().Value(usernameInterface{}).(string)
	path, _ := strings.CutPrefix(r.URL.Path, "/api/reclist/")
	prefix := strings.Split(path, "/")

	api.infoLog.Println(username, len(prefix), prefix)

	if len(prefix) != 0 && prefix[len(prefix)-1] == "" {
		prefix = prefix[:len(prefix)-1]
	}

	listed, err := api.storage.ListEncrypted(username, prefix)
	if err != nil {
		api.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, jsonContentType)
	json.NewEncoder(w).Encode(listed)
}

func (api *API) signUp(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("Sign up")
	var user uData
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil || user.Username == "" || user.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = api.storage.CreateNewUser(user.Username, user.Password)

	if errors.Is(err, encrypt_db.ErrUserAlreadyExist) {
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

// разблокировка с помощью алгорима Шамира

func (api *API) unseal(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("UNSEAL")

	var parts []string

	err := json.NewDecoder(r.Body).Decode(&parts)
	if err != nil {
		api.errorLog.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	decodedData := make([][]byte, len(parts))
	for i, str := range parts {
		api.infoLog.Println(str)
		bytesArray, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			api.errorLog.Println("Ошибка декодирования Base64:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		decodedData[i] = bytesArray
	}

	err = api.storage.InitWrapper(decodedData)
	if err != nil {
		api.errorLog.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	JWTsecretKey, err = api.storage.Get("", nil, JWTname, metaBucketName)
	if err != nil && !errors.Is(encrypt_db.ErrKeyNotFound, err) {
		api.errorLog.Println(err)
		return
	}

	if JWTsecretKey == nil {
		JWTsecretKey, err = encrypt_db.GeneratePassword(12)
		if err != nil {
			return
		}
		err = api.storage.Set("", nil, JWTname, string(JWTsecretKey), metaBucketName)
		if err != nil {
			api.errorLog.Println(err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	api.infoLog.Println("UNSEAL OK")
}

func (api *API) master(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("SEAL")

	var secretInfo encrypt_db.SecretInfo
	err := json.NewDecoder(r.Body).Decode(&secretInfo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	parts, err := encrypt_db.MakeMasterKey(secretInfo)

	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	encodedData := make([]string, len(parts))
	for i, bytesArray := range parts {
		encodedData[i] = base64.StdEncoding.EncodeToString(bytesArray)
	}

	w.Header().Set(headerContentType, jsonContentType)
	json.NewEncoder(w).Encode(encodedData)
	api.infoLog.Println("SEAL OK")
}
