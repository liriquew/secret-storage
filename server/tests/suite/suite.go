package suite

import (
	"fmt"
	"testing"

	"github.com/liriquew/secret_storage/server/internal/lib/config"
)

type Suite struct {
	*testing.T
	TestConfig *config.AppTestConfig
}

func New(t *testing.T) *Suite {
	t.Helper()

	cfg := config.MustLoadTestPath("../config/test_config.yaml")

	return &Suite{
		TestConfig: &cfg,
	}
}

func (s *Suite) GetURL() string {
	return fmt.Sprintf("http://%s:%s/api", s.TestConfig.Service.Host, s.TestConfig.Service.Port)
}
