package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

// type FileStorage struct {
// 	FileJSON *os.File
// 	urlMap   map[string]URL
// }

// func NewFileStorage(f *os.File) (AddorGetURL, error) {
// 	m := make(map[string]URL)
// 	err := json.NewDecoder(f).Decode(&m)
// 	if err != nil && err != io.EOF {
// 		return nil, fmt.Errorf("unable to unmarshal file into map: %w", err)
// 	}
// 	if err == io.EOF {
// 		m = make(map[string]URL)
// 	}
// 	return &FileStorage{
// 		FileJSON: f,
// 		urlMap:   m,
// 	}, nil
// }

// func (fs *FileStorage) AddURL(u URL, userID string) error {
// 	if fs.urlMap == nil {
// 		fs.urlMap = make(map[string]URL)
// 	}
// 	if u.ShortURL == "" {
// 		return ErrKeyNotSpecified
// 	}
// 	fs.urlMap[userID] = u
// 	jsonData, err := json.Marshal(fs.urlMap)
// 	if err != nil {
// 		return fmt.Errorf("unable to marshal internal file storage map: %w", err)
// 	}
// 	if err := os.Truncate("OurURL.json", 0); err != nil {
// 		return fmt.Errorf("unable to Truncate file: %w", err)
// 	}
// 	fs.FileJSON.Write(jsonData)
// 	return nil
// }

// func (fs *FileStorage) GetOriginalURL(_ context.Context, shortID string) (string, error) {
// 	if shortID == "" {
// 		return "", ErrKeyNotSpecified
// 	}
// 	m := make(map[string]string)
// 	for _, v := range fs.urlMap {
// 		m[v.ShortURL] = v.OriginalURL
// 	}
// 	for k, v := range m {
// 		if k == shortID {
// 			return v, nil
// 		}
// 	}
// 	return "", ErrKeyNotFound
// }

// func (fs *FileStorage) GetShortURL(_ context.Context, originalurl string) (string, error) {
// 	if originalurl == "" {
// 		return "", ErrKeyNotSpecified
// 	}
// 	for _, v := range fs.urlMap {
// 		if v.OriginalURL == originalurl {
// 			return v.ShortURL, nil
// 		}
// 	}
// 	return "", ErrKeyNotFound
// }

// func (fs *FileStorage) FindAllUserUrls(_ context.Context, userID string) (map[string]string, error) {
// 	m := make(map[string]string)
// 	for k, v := range fs.urlMap {
// 		if k == userID {
// 			m[v.ShortURL] = v.OriginalURL
// 		}
// 	}
// 	return m, nil
// }

// func (fs *FileStorage) BatchURL(_ctx context.Context, userID string, urls []Correlation) ([]ShortCorrelation, error) {
// 	if fs.urlMap == nil {
// 		fs.urlMap = make(map[string]URL)
// 	}
// 	if err := os.Truncate("OurURL.json", 0); err != nil {
// 		return nil, fmt.Errorf("BatchURL: unable to Truncate file: %w", err)
// 	}
// 	var newurls []ShortCorrelation
// 	var batchurl URL
// 	for _, batch := range urls {
// 		shortID := GenerateRandomString()
// 		shorturl := BaseTest + shortID
// 		newurl := ShortCorrelation{
// 			ShortURL:      shorturl,
// 			CorrelationID: batch.CorrelationID,
// 		}
// 		newurls = append(newurls, newurl)
// 		batchurl = URL{ShortURL: shortID, OriginalURL: batch.OriginalURL}
// 		fs.urlMap[batch.UserID] = batchurl
// 	}
// 	jsonData, err := json.Marshal(fs.urlMap)
// 	if err != nil {
// 		return nil, fmt.Errorf("BatchURL: unable to marshal internal file storage map: %w", err)
// 	}
// 	fs.FileJSON.Write(jsonData)
// 	return newurls, nil
// }

// OLD
type FileStorage struct {
	FileJSON *os.File
	NewMap   map[string]map[string]string
}

func NewFileStorage(f *os.File) (AddorGetURL, error) {
	m := make(map[string]map[string]string)
	err := json.NewDecoder(f).Decode(&m)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("unable to unmarshal file into map: %w", err)
	}
	if errors.Is(err, io.EOF) {
		m = make(map[string]map[string]string)
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

func (fs *FileStorage) GetOriginalURL(_ context.Context, key string) (string, error) {
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

func (fs *FileStorage) GetShortURL(_ context.Context, originalurl string) (string, error) {
	if originalurl == "" {
		return "", ErrKeyNotSpecified
	}
	for _, urlmap := range fs.NewMap {
		if value, ok := urlmap[originalurl]; ok {
			return value, nil
		}
	}
	return "", ErrKeyNotFound
}

func (fs *FileStorage) FindAllUserUrls(_ context.Context, userID string) (map[string]string, error) {
	AllIDUrls, ok := fs.NewMap[userID]
	if !ok {
		return nil, ErrKeyNotFound
	}
	return AllIDUrls, nil
}

func (fs *FileStorage) BatchURL(_ctx context.Context, userID string, urls []Correlation) ([]ShortCorrelation, error) {
	if fs.NewMap == nil {
		fs.NewMap = make(map[string]map[string]string)
	}
	if fs.NewMap[userID] == nil {
		fs.NewMap[userID] = make(map[string]string)
	}
	if err := os.Truncate("OurURL.json", 0); err != nil {
		return nil, fmt.Errorf("BatchURL: unable to Truncate file: %w", err)
	}
	var newurls []ShortCorrelation
	for _, batch := range urls {
		shortID := GenerateRandomString()
		shorturl := BaseTest + shortID
		newurl := ShortCorrelation{
			ShortURL:      shorturl,
			CorrelationID: batch.CorrelationID,
		}
		newurls = append(newurls, newurl)
		fs.NewMap[batch.UserID][shortID] = batch.OriginalURL
	}
	jsonData, err := json.Marshal(fs.NewMap)
	if err != nil {
		return nil, fmt.Errorf("BatchURL: unable to marshal internal file storage map: %w", err)
	}
	fs.FileJSON.Write(jsonData)
	return newurls, nil
}

func (fs *FileStorage) Close() {
	fs.FileJSON.Close()
}
