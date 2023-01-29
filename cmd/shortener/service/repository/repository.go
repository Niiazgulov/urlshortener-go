package repository

import (
	"encoding/json"
	"errors"
	"os"
)

var (
	ErrorKeyNotFound     = errors.New("the key is not found")
	ErrorKeyNotSpecified = errors.New("the key is not specified")
	FileTemp             *os.File
)

type JSONKeymap struct {
	ShortJSON string `json:"result,omitempty"`
	LongJSON  string `json:"url,omitempty"`
}

type URL struct {
	ShortURL    string
	OriginalURL string
}

type AddorGetURL interface {
	AddURL(longandshortURL URL) error
	GetURL(idshortURL string) (string, error)
}

type FileStorage struct {
	allurls  map[string]string
	FileJSON *os.File
}

func NewFileStorage() AddorGetURL {
	return &FileStorage{
		allurls:  make(map[string]string),
		FileJSON: FileTemp,
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
