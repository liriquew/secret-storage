package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	storage "github.com/liriquew/secret_storage/server/internal/encrypted_storage"
	"github.com/liriquew/secret_storage/server/internal/lib/jwt"
	"github.com/liriquew/secret_storage/server/internal/lib/shamir"
	"github.com/liriquew/secret_storage/server/internal/models"
	"github.com/liriquew/secret_storage/server/pkg/logger/sl"
)

func (s *Service) SignUp(c *gin.Context) {
	user := &models.User{}
	if err := c.ShouldBindJSON(user); err != nil {
		c.String(http.StatusBadRequest, "bad json")
		return
	}

	if user.Username == "" || user.Password == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := s.repository.CreateUser(user); err != nil {
		c.Status(http.StatusTeapot)
		return
	}

	token, err := jwt.NewToken(user)
	if err != nil {
		s.log.Error("error while singUp", sl.Err(err))
		c.String(http.StatusInternalServerError, "failed to create jwt")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jwt-token": token,
	})
}

func (s *Service) SignIn(c *gin.Context) {
	s.log.Info("SIGN IN")
	user := &models.User{}
	if err := c.ShouldBindJSON(user); err != nil {
		c.String(http.StatusBadRequest, "bad json")
		return
	}

	if user.Username == "" || user.Password == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := s.repository.CheckUserCredentials(user); err != nil {
		s.log.Error("error while singUp", sl.Err(err))
		c.Status(http.StatusUnauthorized)
		return
	}

	token, err := jwt.NewToken(user)
	if err != nil {
		s.log.Error("error while singUp", sl.Err(err))
		c.String(http.StatusInternalServerError, "failed to create jwt")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jwt-token": token,
	})
}

func (s *Service) Create(c *gin.Context) {
	record := &models.RecordDTO{}
	if err := c.ShouldBindJSON(record); err != nil {
		c.String(http.StatusBadRequest, "bad json")
		return
	}

	path := extractPath(c)

	if err := s.repository.Set(path, record.Key, record.ToInternalRecord().Value); err != nil {
		s.log.Error("error while creating record", sl.Err(err))
		if errors.Is(err, storage.ErrIncorrectPath) {
			c.Status(http.StatusBadRequest)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Service) Get(c *gin.Context) {
	key := c.Param(keyParam)
	path := extractPath(c)

	value, err := s.repository.Get(path, key)
	if err != nil {
		s.log.Error("error while getting record", sl.Err(err))
		if errors.Is(err, storage.ErrEmptyPathPart) {
			c.Status(http.StatusBadRequest)
			return
		}
		if errors.Is(err, storage.ErrBucketNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if errors.Is(err, storage.ErrRecordNotFound) {
			c.Status(http.StatusNotFound)
			return
		}

		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, &models.Record{Key: []byte(key), Value: value})
}

func (s *Service) Update(c *gin.Context) {
	record := &models.RecordDTO{}
	if err := c.ShouldBindJSON(record); err != nil {
		c.String(http.StatusBadRequest, "bad json")
	}

	if record.Key == "" {
		record.Key = c.Param(keyParam)
	}

	path := extractPath(c)

	if err := s.repository.Set(path, record.Key, record.ToInternalRecord().Value); err != nil {
		s.log.Error("error while updating record", sl.Err(err))
		if errors.Is(err, storage.ErrIncorrectPath) {
			c.Status(http.StatusBadRequest)
			return
		}

		c.Status(http.StatusInternalServerError)
	}

	c.Status(http.StatusOK)
}

func (s *Service) Delete(c *gin.Context) {
	key := c.Param(keyParam)
	path := extractPath(c)

	deleted, err := s.repository.Delete(path, key)
	if err != nil {
		s.log.Error("error while getting record", sl.Err(err))
		if errors.Is(err, storage.ErrBucketNotFound) {
			c.Status(http.StatusNotFound)
			return
		}

		c.Status(http.StatusInternalServerError)
	}

	c.JSON(http.StatusOK, gin.H{
		"deleted_buckets": deleted,
	})
}
func (s *Service) ListSecrets(c *gin.Context) {
	path := extractPath(c)

	records, err := s.repository.ListRecords(path)
	if err != nil {
		s.log.Error("error while listing records", sl.Err(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, records)
}

func (s *Service) ListSecretsRecursively(c *gin.Context) {
	path := extractPath(c)

	records, err := s.repository.ListRecordsRecursively(path)
	if err != nil {
		s.log.Error("error while listing recursively records", sl.Err(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, records)
}

func (s *Service) Unseal(c *gin.Context) {
	part := c.Query(partParam)

	if part == "" {
		c.Status(http.StatusBadRequest)
	}

	if err := s.masterKeyInfo.AddPart(part); err != nil {
		if errors.Is(err, shamir.ErrAlreadyAdded) {
			c.String(http.StatusConflict, "part already added")
			return
		}

		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Service) UnsealComplete(c *gin.Context) {
	err := s.Setup()
	if err != nil {
		s.log.Error("error while comleting unseal", sl.Err(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Service) Master(c *gin.Context) {
	threshold, err := strconv.Atoi(c.Query(thresholdParam))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	thresholdOld := s.masterKeyInfo.SetThreshold(threshold)
	if threshold != thresholdOld {
		c.String(http.StatusBadRequest, "thresholds must be equal")
		return
	}

	if err := s.AddConn(c); err != nil {
		s.log.Error("error while adding websocket connection", sl.Err(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Service) MasterComplete(c *gin.Context) {
	err := s.Notify(func(n int) ([][]byte, error) {
		masterKey, err := GeneratePassword(32)
		if err != nil {
			return nil, err
		}
		parts, err := shamir.Split(masterKey, n, s.masterKeyInfo.GetThreshold())
		if err != nil {
			return nil, err
		}
		for i, part := range parts {
			parts[i], _ = json.Marshal(struct {
				Part string `json:"part"`
			}{
				Part: base64.RawStdEncoding.EncodeToString(part),
			})
			s.log.Info("part:", slog.String("part", string(parts[i])))
		}
		return parts, nil
	})
	if err != nil {
		s.log.Error("error while notifing", sl.Err(err))
		c.Status(http.StatusInternalServerError)
	}

	c.Status(http.StatusOK)
}
