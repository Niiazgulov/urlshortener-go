package repository

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/configuration"
)

var (
	ErrKeyNotFound     = errors.New("the key is not found")
	ErrKeyNotSpecified = errors.New("the key is not specified")
	ErrKeyNotExists    = errors.New("the key is not exists")
	ErrIDNotValid      = errors.New("sign-userID is not valid")
)

type AddorGetURL interface {
	AddURL(ctx context.Context, u URL, userID string) error
	GetURL(ctx context.Context, s string) (string, error)
	FindAllUserUrls(ctx context.Context, userID string) (map[string]string, error)
	Close()
}

type JSONKeymap struct {
	ShortJSON string `json:"result,omitempty"`
	LongJSON  string `json:"url,omitempty"`
}

type URL struct {
	ShortURL    string
	OriginalURL string
}

func GetRepository(cfg configuration.Config) (AddorGetURL, error) {
	if cfg.DBPath != "" {
		repo, err := NewDataBaseStorqage(cfg.DBPath)
		if err != nil {
			return nil, fmt.Errorf("GetRepository: unable to make repo (NewDataBaseStorage): %w", err)
		}
		return repo, nil
	} else {
		f, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			return nil, fmt.Errorf("GetRepository: unable to open file: %w", err)
		}
		repo, err := NewFileStorage(f)
		if err != nil {
			return nil, fmt.Errorf("GetRepository: unable to make repo (NewFileStorage): %w", err)
		}
		return repo, nil
	}
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
