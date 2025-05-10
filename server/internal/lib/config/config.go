package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Service ServiceConfig `yaml:"service_config" env-required:"true"`
	Storage StorageConfig `yaml:"storage_config" env-required:"true"`
}

type ServiceConfig struct {
	Port int `yaml:"port" env-required:"true"`
}

type StorageConfig struct {
	Path string `yaml:"path" env-required:"true"`
}

type AppTestConfig struct {
	Service ServiceTestConfig `yaml:"service_config" env-required:"true"`
}

type ServiceTestConfig struct {
	Host string `yaml:"host" env-required:"true"`
	Port string `yaml:"port" env-required:"true"`
}

func MustLoad() AppConfig {
	path := fetchConfigPath()

	if path == "" {
		path = "./config/config.yaml"
	}

	return MustLoadPath(path)
}

func fetchConfigPath() string {
	return os.Getenv("CONF_PATH")
}

func MustLoadPath(path string) AppConfig {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config: file not exist")
	}

	var cfg AppConfig

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("error while reading config" + err.Error())
	}

	return cfg
}

func MustLoadTestPath(path string) AppTestConfig {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config: file not exist")
	}

	var cfg AppTestConfig

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("error while reading config" + err.Error())
	}

	return cfg
}
