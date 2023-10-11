// Пакет Config хранит в себе конфигурацию приложения.
package configuration

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
)

// Структура Config для работы с флагами и переменными окружения.
type Config struct {
	ServerAddress  string `json:"server_address"`
	BaseURLAddress string `json:"base_url"`
	FilePath       string `json:"file_storage_path"`
	DBPath         string `json:"database_path"`
	HTTPS          bool   `json:"enable_https"`
	GRPCTLS        bool   `json:"grpc_enable_tls"`
	ConfigURL      *url.URL
	WorkerCount    int
	CfgFilePath    string
	TrustedSubnet  string `json:"trusted_subnet"`
	ServerGRPC     string `json:"grpc_server_address"`
}

// Функция для создания нового объекта конфигурации.
func NewConfig() (*Config, error) {
	cfg := &Config{
		BaseURLAddress: "",
		ServerAddress:  "",
		ServerGRPC:     "",
		FilePath:       "",
		DBPath:         "",
		HTTPS:          false,
		GRPCTLS:        false,
		WorkerCount:    15,
		CfgFilePath:    "",
		TrustedSubnet:  "",
	}

	flag.StringVar(&cfg.ServerAddress, "a", "", "host to listen on")
	flag.StringVar(&cfg.ServerGRPC, "ga", "", "gRPC server to listen on")
	flag.StringVar(&cfg.BaseURLAddress, "b", "", "base url")
	flag.StringVar(&cfg.FilePath, "f", "", "file storage path")
	flag.StringVar(&cfg.DBPath, "d", "", "database path")
	flag.BoolVar(&cfg.HTTPS, "s", false, "enable https")
	flag.BoolVar(&cfg.GRPCTLS, "gs", false, "gRPC enable https")
	flag.StringVar(&cfg.CfgFilePath, "c", "", "configuration path for file")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "trusted subnet CIDR")
	flag.Parse()
	fileCfg := cfg.readCfgFile(cfg.CfgFilePath)
	cfg.BaseURLAddress = pickFirstNonEmptyString(cfg.BaseURLAddress, os.Getenv("BASE_URL"), fileCfg.BaseURLAddress, "http://localhost:8080")
	cfg.ServerAddress = pickFirstNonEmptyString(cfg.ServerAddress, os.Getenv("SERVER_ADDRESS"), fileCfg.ServerAddress, ":8080")
	cfg.ServerGRPC = pickFirstNonEmptyString(cfg.ServerGRPC, os.Getenv("GRPC_SERVER_ADDRESS"), fileCfg.ServerGRPC, ":3200")
	cfg.FilePath = pickFirstNonEmptyString(cfg.FilePath, os.Getenv("FILE_STORAGE_PATH"), fileCfg.FilePath, "OurURL.json")
	cfg.DBPath = pickFirstNonEmptyString(cfg.DBPath, os.Getenv("DATABASE_DSN"), fileCfg.DBPath)
	cfg.HTTPS = pickFirstNonEmptyBool(cfg.HTTPS, os.Getenv("ENABLE_HTTPS") == "true", fileCfg.HTTPS)
	cfg.GRPCTLS = pickFirstNonEmptyBool(cfg.GRPCTLS, os.Getenv("GRPC_ENABLE_TLS") == "true", fileCfg.GRPCTLS)
	cfg.TrustedSubnet = pickFirstNonEmptyString(cfg.TrustedSubnet, os.Getenv("TRUSTED_SUBNET"), fileCfg.TrustedSubnet, "127.0.0.1/24")
	var err error
	cfg.ConfigURL, err = url.Parse(cfg.BaseURLAddress)
	if err != nil {
		return nil, fmt.Errorf("NewConfig: unable to parse BaseURLAddress: %w", err)
	}

	return cfg, nil
}

// Объект конфига
var Cfg Config

func pickFirstNonEmptyString(strings ...string) string {
	for _, str := range strings {
		if str != "" {
			return str
		}
	}
	return ""
}

func pickFirstNonEmptyBool(bools ...bool) bool {
	for _, boolVar := range bools {
		if boolVar {
			return true
		}
	}
	return false
}

func (c *Config) readCfgFile(cfgPath string) Config {
	fileCfg := Config{}
	if cfgPath == "" {
		return fileCfg
	}
	f, err := os.ReadFile(cfgPath)
	if err != nil {
		return fileCfg
	}
	json.Unmarshal(f, &fileCfg)
	return fileCfg
}
