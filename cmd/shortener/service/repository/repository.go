package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

var (
	ErrKeyNotFound     = errors.New("the key is not found")
	ErrKeyNotSpecified = errors.New("the key is not specified")
	ErrKeyNotExists    = errors.New("the key is not exists")
	ErrSignNotValid    = errors.New("sign is not valid")
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
	AddURL(u URL, userID string) error
	GetURL(s string) (string, error)
	AddNewUser() (string, error)
	FindAllUserUrls(userID string) (map[string]string, error)
}

type FileStorage struct {
	FileJSON *os.File
	// AllUrlsMap  map[string]string
	// UserUrlsMap map[string][]int
	NewMap    map[string]map[string]string
	UserCount int
}

func NewFileStorage(f *os.File) (AddorGetURL, error) {
	m := make(map[string]map[string]string)
	//uum := make(map[string][]int)
	err := json.NewDecoder(f).Decode(&m)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("unable to unmarshal file into map: %w", err)
	}
	if err == io.EOF {
		m = make(map[string]map[string]string)
		//uum = make(map[string][]int)
	}
	return &FileStorage{
		FileJSON: f,
		NewMap:   m,
	}, nil
}

func (fs *FileStorage) AddURL(u URL, userID string) error {
	if fs.NewMap == nil {
		fs.NewMap = make(map[string]map[string]string)
	}
	if fs.NewMap[userID] == nil {
		fs.NewMap[userID] = make(map[string]string)
	}
	if u.ShortURL == "" {
		return ErrKeyNotSpecified
	}
	//fs.NewMap[userID] = make(map[string]string)
	fs.NewMap[userID][u.ShortURL] = u.OriginalURL
	jsonData, err := json.Marshal(fs.NewMap)
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
	for _, urlmap := range fs.NewMap {
		if value, ok := urlmap[key]; ok {
			return value, nil
		}
	}
	return "", ErrKeyNotFound
}

func (fs *FileStorage) AddNewUser() (string, error) {
	newID := GenerateRandomIntString()
	for key := range fs.NewMap {
		if key == newID {
			newID = GenerateRandomIntString()
		}
	}
	return newID, nil
}

func (fs *FileStorage) FindAllUserUrls(userID string) (map[string]string, error) {
	AllIdUrls, checktrue := fs.NewMap[userID]
	if !checktrue {
		return nil, ErrKeyNotFound
	}
	return AllIdUrls, nil
}

const (
	Symbols        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	IntSymbols     = "0123456789"
	ShortURLMaxLen = 7
)

func GenerateRandomString() string {
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, 0, ShortURLMaxLen)
	for i := 0; i < ShortURLMaxLen; i++ {
		s := Symbols[rand.Intn(len(Symbols))]
		result = append(result, s)
	}
	return string(result)
}

func GenerateRandomIntString() string {
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, 0, ShortURLMaxLen)
	for i := 0; i < ShortURLMaxLen; i++ {
		s := IntSymbols[rand.Intn(len(IntSymbols))]
		result = append(result, s)
	}
	return string(result)
}
