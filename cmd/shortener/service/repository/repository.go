package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrKeyNotFound     = errors.New("the key is not found")
	ErrKeyNotSpecified = errors.New("the key is not specified")
	ErrKeyNotExists    = errors.New("the key is not exists")
)

type JSONKeymap struct {
	ShortJSON string `json:"result,omitempty"`
	LongJSON  string `json:"url,omitempty"`
}

type UserURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
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

func NewFileStorage(f *os.File) (AddorGetURL, error) {
	m := make(map[string]string)
	err := json.NewDecoder(f).Decode(&m)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("unable to unmarshal file into map: %w", err)
	}
	if err == io.EOF {
		m = make(map[string]string)
	}
	return &FileStorage{
		allurls:  m,
		FileJSON: f,
	}, nil
}

func (fs *FileStorage) AddURL(u URL) error {
	if fs.allurls == nil {
		fs.allurls = make(map[string]string)
	}
	if u.ShortURL == "" {
		return ErrKeyNotSpecified
	}
	fs.allurls[u.ShortURL] = u.OriginalURL
	jsonData, err := json.Marshal(fs.allurls)
	if err != nil {
		return fmt.Errorf("unable to marshal internal file storage map: %w", err)
	}
	if err := os.Truncate("OurURL.json", 0); err != nil {
		return fmt.Errorf("unable to Truncate file: %w", err)
	}
	fs.FileJSON.Write(jsonData)
	return nil
}

func (fs *FileStorage) GetURL(key string) (string, error) {
	if key == "" {
		return "", ErrKeyNotSpecified
	}
	if value, ok := fs.allurls[key]; ok {
		return value, nil
	}
	return "", ErrKeyNotFound
}
