package configuration

import (
	"flag"
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	ServerAddress  string   `json:"server_address"`
	BaseURLAddress string   `json:"base_url"`
	FilePath       string   `json:"file_storage_path"`
	ConfigURL      *url.URL `json:"config_url"`
	DBPath         string   `json:"database_path"`
	WorkerCount    int
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		BaseURLAddress: "",
		ServerAddress:  "",
		FilePath:       "",
		DBPath:         "",
		WorkerCount:    10,
	}
	flag.StringVar(&cfg.ServerAddress, "a", "", "host to listen on")
	flag.StringVar(&cfg.BaseURLAddress, "b", "", "base url")
	flag.StringVar(&cfg.FilePath, "f", "", "file storage path")
	flag.StringVar(&cfg.DBPath, "d", "", "database path")
	flag.Parse()
	cfg.BaseURLAddress = pickFirstNonEmpty(cfg.BaseURLAddress, os.Getenv("BASE_URL"), "http://localhost:8080")
	cfg.ServerAddress = pickFirstNonEmpty(cfg.ServerAddress, os.Getenv("SERVER_ADDRESS"), ":8080")
	cfg.FilePath = pickFirstNonEmpty(cfg.FilePath, os.Getenv("FILE_STORAGE_PATH"), "OurURL.json")
	cfg.DBPath = pickFirstNonEmpty(cfg.DBPath, os.Getenv("DATABASE_DSN"))
	var err error
	cfg.ConfigURL, err = url.Parse(cfg.BaseURLAddress)
	if err != nil {
		return nil, fmt.Errorf("NewConfig: unable to parse BaseURLAddress: %w", err)
	}
	return cfg, nil
}

var Cfg Config

func pickFirstNonEmpty(strings ...string) string {
	for _, str := range strings {
		if str != "" {
			return str
		}
	}
	return ""
}
