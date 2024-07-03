package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"secret-storage/cli/config"
)

var (
	baseURL string = "http://localhost:8080/api/"
)

func prepareRequest(method, path string, data *KV) (*http.Response, error) {
	token := config.GetToken()

	client := &http.Client{}

	if len(path) != 0 && path[len(path)-1] != '/' {
		path += "/"
	}

	var key string
	var buf []byte
	if data != nil {
		key = data.Key

		var err error
		buf, err = json.Marshal(&data)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, baseURL+path+key, bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}
