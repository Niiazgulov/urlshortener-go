package configuration

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress  string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURLAddress string `env:"BASE_URL" envDefault:"http://localhost:8080/"`
	FilePath       string `env:"FILE_STORAGE_PATH"`
}

var (
	Cfg Config

// FlagServer string
// FlagBase   string
// FlagFile   string
)

func ChoosePriority(strings ...string) string {
	for _, str := range strings {
		if str != "" {
			return str
		}
	}
	return ""
}

func MakeConfig() (*Config, error) {
	Cfg := &Config{
		BaseURLAddress: "",
		ServerAddress:  "",
		FilePath:       "",
	}
	flag.StringVar(&Cfg.ServerAddress, "a", "", "server adress")
	flag.StringVar(&Cfg.BaseURLAddress, "b", "", "base url adress")
	flag.StringVar(&Cfg.FilePath, "f", "", "file path")
	flag.Parse()
	Cfg.BaseURLAddress = ChoosePriority(Cfg.BaseURLAddress, os.Getenv("BASE_URL"), "http://localhost:8080")
	Cfg.ServerAddress = ChoosePriority(Cfg.ServerAddress, os.Getenv("SERVER_ADDRESS"), ":8080")
	Cfg.FilePath = ChoosePriority(Cfg.FilePath, os.Getenv("FILE_STORAGE_PATH"))
	return Cfg, nil
}

func MakeCfgVars(ba, sa, fp string) {
	flag.StringVar(&sa, "a", "", "server adress")
	flag.StringVar(&ba, "b", "", "base url adress")
	flag.StringVar(&fp, "f", "", "file path")
	flag.Parse()
	ba = ChoosePriority(ba, os.Getenv("BASE_URL"), "http://localhost:8080")
	sa = ChoosePriority(sa, os.Getenv("SERVER_ADDRESS"), ":8080")
	fp = ChoosePriority(fp, os.Getenv("FILE_STORAGE_PATH"))
}
