package repository

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/configuration"
)

var (
	ErrorKeyNotFound     = errors.New("the key is not found")
	ErrorKeyNotSpecified = errors.New("the key is not specified")
	FileTemp             *os.File
	err                  error
)

// func init() {

// 	FileTemp, err = os.OpenFile(configuration.Cfg.FilePath, os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
// 	if err != nil {
// 		panic(err)
// 	}
// }

type URL struct {
	ShortURL    string
	OriginalURL string
}

// type MemoryRepository struct {
// 	allurls map[string]string
// }

type AddorGetURL interface {
	AddURL(longandshortURL URL) error
	GetURL(idshortURL string) (string, error)
}

// func NewMemoryRepository() AddorGetURL {
// 	return &MemoryRepository{
// 		allurls: make(map[string]string),
// 	}
// }

// func (mr *MemoryRepository) AddURL(u URL) error {
// 	if mr.allurls == nil {
// 		mr.allurls = make(map[string]string)
// 	}
// 	if u.ShortURL == "" {
// 		return ErrorKeyNotSpecified
// 	}
// 	mr.allurls[u.ShortURL] = u.OriginalURL
// 	return nil
// }

// func (mr *MemoryRepository) GetURL(key string) (string, error) {
// 	if key == "" {
// 		return "", ErrorKeyNotSpecified
// 	}
// 	if value, ok := mr.allurls[key]; ok {
// 		return value, nil
// 	}
// 	return "", ErrorKeyNotFound
// }

type FileStorage struct {
	allurls  map[string]string
	FileJSON *os.File
}

func NewFileStorage() AddorGetURL {
	return &FileStorage{
		allurls: make(map[string]string),
		// FileJSON: FileTemp,
	}
}

func (fs *FileStorage) AddURL(u URL) error {
	if fs.allurls == nil {
		fs.allurls = make(map[string]string)
	}
	if u.ShortURL == "" {
		return ErrorKeyNotSpecified
	}
	fs.allurls[u.ShortURL] = u.OriginalURL
	// var err error
	fs.FileJSON, err = os.OpenFile(configuration.Cfg.FilePath, os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		panic(err)
	}
	jsonData, err := json.Marshal(fs.allurls)
	if err != nil {
		panic(err)
	}
	fs.FileJSON.Write(jsonData)
	defer fs.FileJSON.Close()
	return nil
}

func (fs *FileStorage) GetURL(key string) (string, error) {
	if key == "" {
		return "", ErrorKeyNotSpecified
	}
	if value, ok := fs.allurls[key]; ok {
		return value, nil
	}
	return "", ErrorKeyNotFound
}

// type AddGetFileInterf interface {
// 	AddURL(fileadress string, longandshortURL URL) error
// 	GetURL(idshortURL string) (string, error)
// }

// func NewFileStorage() AddGetFileInterf {
// 	return &FileStorage{
// 		allurls:  make(map[string]string),
// 		FileJSON: FileTemp,
// 	}
// }
