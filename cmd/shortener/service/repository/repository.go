package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/configuration"
)

var (
	ErrKeyNotFound     = errors.New("the key is not found")
	ErrKeyNotSpecified = errors.New("the key is not specified")
	ErrKeyNotExists    = errors.New("the key is not exists")
	ErrIDNotValid      = errors.New("sign-userID is not valid")
	ErrURLexists       = errors.New("such URl already exist in DB")
)

type AddorGetURL interface {
	AddURL(u URL, userID string) error
	GetOriginalURL(ctx context.Context, s string) (string, error)
	GetShortURL(ctx context.Context, s string) (string, error)
	FindAllUserUrls(ctx context.Context, userID string) (map[string]string, error)
	BatchURL(ctx context.Context, userID string, originalurls []Correlation) ([]ShortCorrelation, error)
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

const (
	Symbols        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	IntSymbols     = "0123456789"
	ShortURLMaxLen = 7
	BaseTest       = "http://localhost:8080/"
)

func GetRepository(cfg *configuration.Config) (AddorGetURL, error) {
	if cfg.DBPath != "" {
		repo, err := NewDataBaseStorqage(cfg.DBPath)
		if err != nil {
			return nil, fmt.Errorf("GetRepository: unable to make repo (NewDataBaseStorage): %w", err)
		}
		return repo, nil
	}
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

func GenerateRandomString() string {
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, 0, ShortURLMaxLen)
	for i := 0; i < ShortURLMaxLen; i++ {
		s := Symbols[rand.Intn(len(Symbols))]
		result = append(result, s)
	}
	return string(result)
}

func RandomStringUniqueCheck(repo AddorGetURL, w http.ResponseWriter, r *http.Request, shortID string) string {
	for {
		if _, err := repo.GetOriginalURL(r.Context(), shortID); err != nil {
			if err == ErrKeyNotFound {
				break
			} else {
				log.Printf("PostHandler: unable to get URL by short ID: %v", err)
				http.Error(w, "PostHandler: unable to get url from DB", http.StatusInternalServerError)
				return ""
			}
		}
		shortID = GenerateRandomString()
	}
	return shortID
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

type ShortCorrelation struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type Correlation struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
	UserID        string `json:"user_id"`
}
