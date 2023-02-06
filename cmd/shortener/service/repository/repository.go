package repository

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

var (
	ErrKeyNotFound     = errors.New("the key is not found")
	ErrKeyNotSpecified = errors.New("the key is not specified")
	ErrKeyNotExists    = errors.New("the key is not exists")
	ErrIDNotValid      = errors.New("sign-userID is not valid")
)

type AddorGetURL interface {
	AddURL(u URL, userID string) error
	GetURL(ctx context.Context, s string) (string, error)
	FindAllUserUrls(ctx context.Context, userID string) (map[string]string, error)
	BatchURL(ctx context.Context, userID string, originalurls []Correlation) ([]Correlation, error)
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

type ShortCorrelation struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type Correlation struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
	ShortURL      string `json:"short_url"`
	UserID        string `json:"user_id"`
}
