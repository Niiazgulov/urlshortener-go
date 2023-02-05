package configuration

import (
	"flag"
	"fmt"
	"net/url"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	ServerAddress  string   `json:"server_address"`
	BaseURLAddress string   `json:"base_url"`
	FilePath       string   `json:"file_storage_path"`
	ConfigURL      *url.URL `json:"config_url"`
	DBPath         string   `json:"database_path"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		BaseURLAddress: "",
		ServerAddress:  "",
		FilePath:       "",
		DBPath:         "",
	}
	flag.StringVar(&cfg.ServerAddress, "a", "", "host to listen on")
	flag.StringVar(&cfg.BaseURLAddress, "b", "", "base url")
	flag.StringVar(&cfg.FilePath, "f", "", "file storage path")
	flag.StringVar(&cfg.DBPath, "d", "", "database path")
	flag.Parse()
	user := "postgres"
	password := "180612"
	dbname := "urldb"
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)
	cfg.BaseURLAddress = pickFirstNonEmpty(cfg.BaseURLAddress, os.Getenv("BASE_URL"), "http://localhost:8080")
	cfg.ServerAddress = pickFirstNonEmpty(cfg.ServerAddress, os.Getenv("SERVER_ADDRESS"), ":8080")
	cfg.FilePath = pickFirstNonEmpty(cfg.FilePath, os.Getenv("FILE_STORAGE_PATH"), "OurURL.json")
	cfg.DBPath = pickFirstNonEmpty(cfg.DBPath, os.Getenv("DATABASE_DSN"), connectionString)
	var err error
	cfg.ConfigURL, err = url.Parse(cfg.BaseURLAddress)
	if err != nil {
		return nil, fmt.Errorf("NewConfig: unable to parse BaseURLAddress: %w", err)
	}
	// cfg.FileTemp, err = os.OpenFile(Cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	// if err != nil {
	// 	return nil, fmt.Errorf("NewConfig: unable to open File: %w", err)
	// }
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
