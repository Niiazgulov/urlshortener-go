// Пакет repository, описание в файле doc.go
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

	"github.com/Niiazgulov/urlshortener-go.git/internal/configuration"
)

// Основные ошибки проекта, используемые для работы с хранилищем.
var (
	ErrKeyNotFound     = errors.New("the key is not found")
	ErrKeyNotSpecified = errors.New("the key is not specified")
	ErrKeyNotExists    = errors.New("the key is not exists")
	ErrIDNotValid      = errors.New("sign-userID is not valid")
	ErrURLexists       = errors.New("such URl already exist in DB")
	ErrURLdeleted      = errors.New("URl previosly deleted")
)

//go:generate mockgen -source=repository.go -destination=mocks/mock_repo.go

// Основной интерфейс проекта, в котором описаны методы работы с хранилищем.
type AddorGetURL interface {
	AddURL(u URL) error
	GetOriginalURL(ctx context.Context, s string) (string, error)
	GetShortURL(ctx context.Context, s string) (string, error)
	FindAllUserUrls(ctx context.Context, userID string) (map[string]string, error)
	BatchURL(ctx context.Context, originalurls []URL) ([]ShortCorrelation, error)
	// BatchURL(ctx context.Context, userID string, originalurls []URL) ([]ShortCorrelation, error)
	DeleteUrls([]URL) error
	Close()
}

// Основная структура проекта, в котором описаны все поля с используемой информацией.
type URL struct {
	ShortURL      string `json:"id"`
	OriginalURL   string `json:"original_url"`
	UserID        string `json:"user_id"`
	CorrelationID string `json:"correlation_id"`
	Deleted       bool   `json:"deleted"`
}

// Структура для хэндлера удаления урлов.
type DeleteURLsJob struct {
	RequestURLs []URL
	UserID      string
}

// Структура для хэндлера POST для работы с JSON.
type JSONKeymap struct {
	ShortJSON string `json:"result,omitempty"`
	LongJSON  string `json:"url,omitempty"`
}

// Структура для хэндлера PostBatchHandler.
type ShortCorrelation struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// Используемые константы.
const (
	Symbols        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	IntSymbols     = "0123456789"
	ShortURLMaxLen = 7
	BaseTest       = "http://localhost:8080/"
)

// Функция оболочка для определения типа хранилища в зависимости от выданного системой флага.
func GetRepository(cfg *configuration.Config) (AddorGetURL, error) {
	if cfg.DBPath != "" {
		repo, err := NewDataBaseStorage(cfg.DBPath)
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

// Функция-генератор псведо-случайной строки.
func GenerateRandomString() string {
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, 0, ShortURLMaxLen)
	for i := 0; i < ShortURLMaxLen; i++ {
		s := Symbols[rand.Intn(len(Symbols))]
		result = append(result, s)
	}
	return string(result)
}

// Функция для проверки строки на уникальность.
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

// Функция генератор псведо-случайной строки из цифр.
func GenerateRandomIntString() string {
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, 0, ShortURLMaxLen)
	for i := 0; i < ShortURLMaxLen; i++ {
		s := IntSymbols[rand.Intn(len(IntSymbols))]
		result = append(result, s)
	}
	return string(result)
}
