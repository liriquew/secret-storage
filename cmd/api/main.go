package main

import (
	"kv-storage/pkg/bolt"
	"log"
	"net/http"
	"os"
)

const (
	db_path = "./data.db"
)

type API struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	storage  *bolt.Bolt
}

var JWTsecretKey []byte

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ltime)

	storage, err := bolt.New(db_path)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer storage.Close()

	JWTsecretKey = []byte("JWT_SECRET")

	app := API{
		infoLog:  infoLog,
		errorLog: errorLog,
		storage:  storage,
	}

	mux := app.routes()

	app.infoLog.Printf("Starting server on port %s", os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(":8080", mux))
}
