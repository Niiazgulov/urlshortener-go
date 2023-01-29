package configuration

import (
	"flag"
	"net/url"
	"os"
)

type Config struct {
	ServerAddress  string   `json:"server_address"`
	BaseURLAddress string   `json:"base_url"`
	FilePath       string   `json:"file_storage_path"`
	ConfigURL      *url.URL `json:"config_url"`
}

var Cfg Config

func NewConfig() (*Config, error) {
	cfg := &Config{
		BaseURLAddress: "",
		ServerAddress:  "",
		FilePath:       "",
	}
	flag.StringVar(&cfg.ServerAddress, "a", "", "host to listen on")
	flag.StringVar(&cfg.BaseURLAddress, "b", "", "base url")
	flag.StringVar(&cfg.FilePath, "f", "", "file storage path")
	flag.Parse()
	cfg.BaseURLAddress = pickFirstNonEmpty(cfg.BaseURLAddress, os.Getenv("BASE_URL"), "http://localhost:8080")
	cfg.ServerAddress = pickFirstNonEmpty(cfg.ServerAddress, os.Getenv("SERVER_ADDRESS"), ":8080")
	cfg.FilePath = pickFirstNonEmpty(cfg.FilePath, os.Getenv("FILE_STORAGE_PATH"), "OurURL.json")
	cfg.ConfigURL, _ = url.Parse(cfg.BaseURLAddress)
	return cfg, nil
}

func pickFirstNonEmpty(strings ...string) string {
	for _, str := range strings {
		if str != "" {
			return str
		}
	}
	return ""
}
