package tests

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/liriquew/secret_storage/server/internal/models"
	"github.com/liriquew/secret_storage/server/tests/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GetRandRecord() *models.RecordDTO {
	return &models.RecordDTO{
		Key:    gofakeit.Book().Author,
		Value:  gofakeit.Book().Title,
		Base64: false,
	}
}

func GetRandBytes(n int) []byte {
	b := make([]byte, n)
	for i := range n {
		b[i] = byte(rand.Int() % 256)
	}
	return b
}

func GetRandRecordBase64() *models.RecordDTO {
	key := base64.StdEncoding.EncodeToString(GetRandBytes(10))
	value := base64.StdEncoding.EncodeToString(GetRandBytes(10))
	return &models.RecordDTO{
		Key:    key,
		Value:  value,
		Base64: true,
	}
}

var (
	StatusNotFound = "404 Not Found"
)

func CreateRecord(t *testing.T, ts *suite.Suite, userCreds *UserWithToken, path string, record *models.RecordDTO) *models.RecordDTO {
	if record == nil {
		record = GetRandRecord()
	}
	buf, _ := json.Marshal(record)

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/secrets?path=%s", ts.GetURL(), path), bytes.NewBuffer(buf))
	req.Header.Set(contentType, applicationJSON)
	req.Header.Set("Authorization", "Bearer "+userCreds.Token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	return record
}

func GetRecord(t *testing.T, ts *suite.Suite, userCreds *UserWithToken, key, path string) *models.RecordDTO {
	req, _ := http.NewRequest("GET",
		fmt.Sprintf("%s/secrets/%s?path=%s", ts.GetURL(), key, path),
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+userCreds.Token)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, StatusOK, resp.Status)

	record := &models.RecordDTO{}
	json.NewDecoder(resp.Body).Decode(&record)

	return record
}

func TestCreate(t *testing.T) {
	ts := suite.New(t)

	userCreds := CreateUser(t, ts)

	record := GetRandRecord()

	t.Run("Common Path", func(t *testing.T) {
		buf, _ := json.Marshal(record)

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/secrets", ts.GetURL()), bytes.NewBuffer(buf))
		req.Header.Set(contentType, applicationJSON)
		req.Header.Set("Authorization", "Bearer "+userCreds.Token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Some Path", func(t *testing.T) {
		buf, _ := json.Marshal(record)

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/secrets?path=123/123", ts.GetURL()), bytes.NewBuffer(buf))
		req.Header.Set(contentType, applicationJSON)
		req.Header.Set("Authorization", "Bearer "+userCreds.Token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

}

func TestGet(t *testing.T) {
	ts := suite.New(t)

	userCreds := CreateUser(t, ts)

	t.Run("Common Record Success", func(t *testing.T) {
		record := CreateRecord(t, ts, userCreds, "", nil)

		req, _ := http.NewRequest("GET",
			fmt.Sprintf("%s/secrets/%s", ts.GetURL(), record.Key),
			nil,
		)
		req.Header.Set("Authorization", "Bearer "+userCreds.Token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, StatusOK, resp.Status)

		recordFromReq := &models.RecordDTO{}
		json.NewDecoder(resp.Body).Decode(&recordFromReq)

		assert.Equal(t, record.Key, recordFromReq.Key)
		assert.Equal(t, record.Value, recordFromReq.Value)
	})

	t.Run("Base64 record Success", func(t *testing.T) {
		record := CreateRecord(t, ts, userCreds, "", GetRandRecordBase64())

		req, _ := http.NewRequest("GET",
			fmt.Sprintf("%s/secrets/%s", ts.GetURL(), record.Key),
			nil,
		)
		req.Header.Set("Authorization", "Bearer "+userCreds.Token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, StatusOK, resp.Status)

		recordFromReq := &models.RecordDTO{}
		json.NewDecoder(resp.Body).Decode(&recordFromReq)

		assert.Equal(t, record.Key, recordFromReq.Key)
		assert.Equal(t, record.Value, recordFromReq.Value)
	})

	t.Run("Some Path", func(t *testing.T) {
		path := "123/path/some"
		record := CreateRecord(t, ts, userCreds, path, GetRandRecordBase64())

		req, _ := http.NewRequest("GET",
			fmt.Sprintf("%s/secrets/%s?path=%s", ts.GetURL(), record.Key, path),
			nil,
		)
		req.Header.Set("Authorization", "Bearer "+userCreds.Token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, StatusOK, resp.Status)

		recordFromReq := &models.RecordDTO{}
		json.NewDecoder(resp.Body).Decode(&recordFromReq)

		assert.Equal(t, record.Key, recordFromReq.Key)
		assert.Equal(t, record.Value, recordFromReq.Value)
	})

	t.Run("Not found", func(t *testing.T) {
		CreateRecord(t, ts, userCreds, "", nil)
		req, _ := http.NewRequest("GET",
			fmt.Sprintf("%s/secrets/%s?path=%s", ts.GetURL(), "abra_cadabra", "path_does_not_exists"),
			nil,
		)
		req.Header.Set("Authorization", "Bearer "+userCreds.Token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, StatusNotFound, resp.Status)
	})
}

func TestUpdate(t *testing.T) {
	ts := suite.New(t)
	userCreds := CreateUser(t, ts)

	t.Run("Success Update", func(t *testing.T) {
		record := CreateRecord(t, ts, userCreds, "", nil)

		newRecord := models.RecordDTO{
			Value: "123",
		}

		buf, err := json.Marshal(newRecord)

		req, _ := http.NewRequest("PATCH", fmt.Sprintf("%s/secrets/%s", ts.GetURL(), record.Key), bytes.NewBuffer(buf))
		req.Header.Set("Authorization", "Bearer "+userCreds.Token)
		req.Header.Set(contentType, applicationJSON)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, StatusOK, resp.Status)

		updatedRecord := GetRecord(t, ts, userCreds, record.Key, "")

		assert.Equal(t, record.Key, updatedRecord.Key)
		assert.Equal(t, newRecord.Value, updatedRecord.Value)
	})

	t.Run("Success Update With Path", func(t *testing.T) {
		path := "path/to/value"
		record := CreateRecord(t, ts, userCreds, path, nil)

		newRecord := models.RecordDTO{
			Value: "123",
		}

		buf, err := json.Marshal(newRecord)

		req, _ := http.NewRequest("PATCH", fmt.Sprintf("%s/secrets/%s?path=%s", ts.GetURL(), record.Key, path), bytes.NewBuffer(buf))
		req.Header.Set("Authorization", "Bearer "+userCreds.Token)
		req.Header.Set(contentType, applicationJSON)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, StatusOK, resp.Status)

		updatedRecord := GetRecord(t, ts, userCreds, record.Key, path)

		assert.Equal(t, record.Key, updatedRecord.Key)
		assert.Equal(t, newRecord.Value, updatedRecord.Value)
	})
}

func TestDelete(t *testing.T) {
	ts := suite.New(t)

	userCreds := CreateUser(t, ts)

	t.Run("Success Delete", func(t *testing.T) {
		record := CreateRecord(t, ts, userCreds, "", nil)

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/secrets/%s", ts.GetURL(), record.Key), nil)
		req.Header.Set("Authorization", "Bearer "+userCreds.Token)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, StatusOK, resp.Status)
	})

	t.Run("Success Delete With Path", func(t *testing.T) {
		path := "path/to/value"

		record := CreateRecord(t, ts, userCreds, path, nil)

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/secrets/%s?path=%s", ts.GetURL(), record.Key, path), nil)
		req.Header.Set("Authorization", "Bearer "+userCreds.Token)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, StatusOK, resp.Status)
	})

	t.Run("Key does not existst", func(t *testing.T) {
		path := "path/to/value"

		key := "123"

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/secrets/%s?path=%s", ts.GetURL(), key, path), nil)
		req.Header.Set("Authorization", "Bearer "+userCreds.Token)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, "404 Not Found", resp.Status)
	})
	t.Run("Path does not existst", func(t *testing.T) {
		path := "123123113/123213123/to/value"

		key := "123"

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/secrets/%s?path=%s", ts.GetURL(), key, path), nil)
		req.Header.Set("Authorization", "Bearer "+userCreds.Token)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, "404 Not Found", resp.Status)
	})
}

func TestListSecrets(t *testing.T) {
	ts := suite.New(t)

	userCreds := CreateUser(t, ts)

	paths := []string{
		"123/321",
		"123/a/b/c",
		"123/a/b/c",
	}

	wg := sync.WaitGroup{}
	wg.Add(len(paths))
	for i := range paths {
		go func() {
			defer wg.Done()
			CreateRecord(t, ts, userCreds, paths[i], nil)
		}()
	}
	wg.Wait()

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/list", ts.GetURL()), nil)
	req.Header.Set("Authorization", "Bearer "+userCreds.Token)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var info models.BucketInfo

	err = json.NewDecoder(resp.Body).Decode(&info)
	assert.NoError(t, err)

	assert.Len(t, info.Buckets, 1)
	assert.Len(t, info.Records, 0)
}

func TestListSecretsRecursively(t *testing.T) {
	ts := suite.New(t)

	userCreds := CreateUser(t, ts)

	paths := []string{
		"123/321",
		"123/a/b/c",
		"123/a/b/c",
	}

	wg := sync.WaitGroup{}
	wg.Add(len(paths))
	for i := range paths {
		go func() {
			defer wg.Done()
			CreateRecord(t, ts, userCreds, paths[i], nil)
		}()
	}
	wg.Wait()

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/reclist", ts.GetURL()), nil)
	req.Header.Set("Authorization", "Bearer "+userCreds.Token)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var info models.BucketFullInfo

	err = json.NewDecoder(resp.Body).Decode(&info)
	assert.NoError(t, err)

	assert.Len(t, info.Buckets, 1)
	assert.Len(t, info.Records, 0)
}
