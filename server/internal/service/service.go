package service

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	encryptedstorage "github.com/liriquew/secret_storage/server/internal/encrypted_storage"
	"github.com/liriquew/secret_storage/server/internal/lib/config"
	"github.com/liriquew/secret_storage/server/internal/lib/shamir"
	"github.com/liriquew/secret_storage/server/internal/models"
	socketnotifier "github.com/liriquew/secret_storage/server/internal/socket_notifier"
)

type Storage interface {
	Set(path []string, key string, value []byte) error
	Get(path []string, key string) ([]byte, error)
	Delete(path []string, key string) (int, error)
	ListRecords(path []string) (*models.BucketInfo, error)
	ListRecordsRecursively(path []string) (*models.BucketFullInfo, error)

	CreateUser(user *models.User) error
	CheckUserCredentials(user *models.User) error
}

var (
	pathParam      = "path"
	partParam      = "part"
	thresholdParam = "threshold"

	usernameKey = "username"

	keyParam = "key"
)

type Service struct {
	repository Storage
	log        *slog.Logger

	*socketnotifier.Notifier
	masterKeyInfo shamir.ShamirInfo
	storageCfg    config.StorageConfig
}

func New(log *slog.Logger, storageConfig config.StorageConfig) *Service {
	return &Service{
		log:           log,
		masterKeyInfo: shamir.NewShamirInfo(),
		storageCfg:    storageConfig,
		Notifier:      socketnotifier.New(log),
	}
}

func (s *Service) IsReady(c *gin.Context) {
	if s.repository == nil {
		c.Status(http.StatusExpectationFailed)
	} else {
		c.Status(http.StatusOK)
	}
}

func (s *Service) Setup() error {
	defer s.masterKeyInfo.Reset()
	parts, err := s.masterKeyInfo.Parts()
	if err != nil {
		return err
	}

	secret, err := shamir.Combine(parts)
	if err != nil {
		return err
	}

	storage, err := encryptedstorage.New(s.storageCfg, secret)
	if err != nil {
		return err
	}

	s.repository = storage
	return nil
}
