package main

import (
	"encoding/json"
	"net/http"
)

type kvData struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

const (
	headerContentType = "Content-Type"
	jsonContentType   = "application/json"
)

func (app *App) getByKey(w http.ResponseWriter, r *http.Request) {
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

func (app *App) setValueByKey(w http.ResponseWriter, r *http.Request) {
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

func (app *App) deleteByKey(w http.ResponseWriter, r *http.Request) {
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

func (app *App) test(w http.ResponseWriter, r *http.Request) {
	app.storage.List()
}
