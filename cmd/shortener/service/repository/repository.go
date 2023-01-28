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

type URL struct {
	ShortURL    string
	OriginalURL string
}

// type MemoryRepository struct {
// 	allurls map[string]string
// }

// type AddorGetURL interface {
// 	AddURL(longandshortURL URL) error
// 	GetURL(idshortURL string) (string, error)
// }

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

type AddGetFileInterf interface {
	AddURL(fileadress string, longandshortURL URL) error
	GetURL(idshortURL string) (string, error)
}

func NewFileStorage() AddGetFileInterf {
	return &FileStorage{
		allurls:  make(map[string]string),
		FileJSON: FileTemp,
	}
}

func (fs *FileStorage) AddURL(f string, u URL) error {
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
