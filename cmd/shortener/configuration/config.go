package configuration

// import "flag"

type Config struct {
	ServerAddress  string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURLAddress string `env:"BASE_URL" envDefault:"http://localhost:8080/"`
	FilePath       string `env:"FILE_STORAGE_PATH"`
}

var Cfg Config
