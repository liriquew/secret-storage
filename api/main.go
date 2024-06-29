package main

import (
	"log"
	"net/http"
	"os"
	"secret-storage/storage/encrypt_db"
)

type API struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	storage  *encrypt_db.BoltEncrypt
}

var JWTsecretKey []byte

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ltime)

	app := API{
		infoLog:  infoLog,
		errorLog: errorLog,
	}

	mux := app.routes()

	app.infoLog.Printf("Запуск сервера на порте: :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
