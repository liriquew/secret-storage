package models

import (
	"encoding/base64"
	"encoding/json"
)

type RecordDTO struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Base64 bool   `json:"is_base64"`
}

const (
	Base64Encoding = 1
	UTF8Encoding   = 0
)

func (r *RecordDTO) ToInternalRecord() *Record {
	var key, value []byte
	if r.Base64 {
		key, _ = base64.StdEncoding.DecodeString(r.Key)
		value, _ = base64.StdEncoding.DecodeString(r.Value)
		key = append(key, Base64Encoding)
		value = append(value, Base64Encoding)
	} else {
		key, value = []byte(r.Key), []byte(r.Value)
		key = append(key, UTF8Encoding)
		value = append(value, UTF8Encoding)
	}

	return &Record{
		Key:   key,
		Value: value,
	}
}

type Record struct {
	Key   []byte
	Value []byte
}

func (r *Record) MarshalJSON() ([]byte, error) {
	var value string

	if n := len(r.Value); n > 0 && r.Value[n-1] == UTF8Encoding {
		value = string(r.Value[:n-1])
	} else {
		value = base64.StdEncoding.EncodeToString(r.Value[:n-1])
	}

	return json.Marshal(RecordDTO{
		Key:   string(r.Key),
		Value: value,
	})
}

func (r *Record) UnmarshalJSON(data []byte) error {
	dto := RecordDTO{}

	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}

	r.Key = []byte(dto.Key)

	if dto.Base64 {
		decoded, err := base64.StdEncoding.DecodeString(dto.Value)
		if err != nil {
			return err
		}
		r.Value = decoded
	} else {
		r.Value = []byte(dto.Value)
	}

	return nil
}

type BucketInfo struct {
	Buckets []string  `json:"buckets"`
	Records []*Record `json:"records"`
}

type BucketFullInfo struct {
	Name    string            `json:"name"`
	Buckets []*BucketFullInfo `json:"buckets"`
	Records []*Record         `json:"records"`
}
