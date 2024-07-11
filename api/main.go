package main

import (
	"encoding/base64"
	"flag"
	"fmt"
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

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ltime)

	MakeMasterKey()

	dbInfo, err := encrypt_db.NewEncryptKV(dbPath)
	if err != nil {
		errorLog.Println(err)
		return
	}

	api := API{
		infoLog:  infoLog,
		errorLog: errorLog,
		storage:  dbInfo.Storage,
	}

	api.infoLog.Printf("RootName: root")
	api.infoLog.Printf("RootPass: %s", dbInfo.RootPass)

	mux := api.routes()

	api.infoLog.Printf("Запуск сервера на порте: :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func MakeMasterKey() {
	parts := flag.Int("p", -1, "Общее число частей")
	threshold := flag.Int("t", -1, "Число частей для разблокировки")
	flag.Parse()

	if *parts == -1 || *threshold == -1 {
		return
	}

	keyParts, err := encrypt_db.MakeMasterKey(encrypt_db.SecretInfo{Parts: *parts, Threshold: *threshold})
	if err != nil {
		fmt.Println(err)
		return
	}

	for i, keyPart := range keyParts {
		fmt.Printf("%d:\t%s\n", i+1, string(base64.StdEncoding.EncodeToString(keyPart)))
	}
}
