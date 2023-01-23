package configuration

import (
	"flag"
	"os"
)

type Config struct {
	// ServerAddress  string `env:"SERVER_ADDRESS" envDefault:":8080"`
	// BaseURLAddress string `env:"BASE_URL" envDefault:"http://localhost:8080/"`
	// FilePath       string `env:"FILE_STORAGE_PATH"`
	ServerAddress  string `json:"server_address"`
	BaseURLAddress string `json:"base_url"`
	FilePath       string `json:"file_storage_path"`
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
	cfg.BaseURLAddress = ChoosePriority(cfg.BaseURLAddress, os.Getenv("BASE_URL"), "http://localhost:8080")
	cfg.ServerAddress = ChoosePriority(cfg.ServerAddress, os.Getenv("SERVER_ADDRESS"), ":8080")
	cfg.FilePath = ChoosePriority(cfg.FilePath, os.Getenv("FILE_STORAGE_PATH"))
	return cfg, nil
}

func ChoosePriority(strings ...string) string {
	for _, str := range strings {
		if str != "" {
			return str
		}
	}
	return ""
}
