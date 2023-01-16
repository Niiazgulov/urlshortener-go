package repository

import (
	"errors"
)

var (
	ErrorKeyNotFound     = errors.New("the key is not found")
	ErrorKeyNotSpecified = errors.New("the key is not specified")
)

type URL struct {
	ShortURL    string
	OriginalURL string
}

type MemoryRepository struct {
	allurls map[string]string
}

type AddorGetURL interface {
	AddURL(longandshortURL URL) error
	GetURL(idshortURL string) (string, error)
}

func NewMemoryRepository() AddorGetURL {
	return &MemoryRepository{
		allurls: make(map[string]string),
	}
}

func (mr *MemoryRepository) AddURL(u URL) error {
	if mr.allurls == nil {
		mr.allurls = make(map[string]string)
	}
	if u.ShortURL == "" {
		return ErrorKeyNotSpecified
	}
	mr.allurls[u.ShortURL] = u.OriginalURL
	return nil
}

func (mr *MemoryRepository) GetURL(key string) (string, error) {
	if key == "" {
		return "", ErrorKeyNotSpecified
	}
	if value, ok := mr.allurls[key]; ok {
		return value, nil
	}
	return "", ErrorKeyNotFound
}
