package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type FileStorage struct {
	FileJSON *os.File
	NewMap   map[string]map[string]string
	Writer   *bufio.Writer
}

func NewFileStorage(f *os.File) (AddorGetURL, error) {
	m := make(map[string]map[string]string)
	err := json.NewDecoder(f).Decode(&m)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("unable to unmarshal file into map: %w", err)
	}
	if err == io.EOF {
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

func (fs *FileStorage) GetURL(_ context.Context, key string) (string, error) {
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

func (fs *FileStorage) FindAllUserUrls(_ context.Context, userID string) (map[string]string, error) {
	AllIDUrls, ok := fs.NewMap[userID]
	if !ok {
		return nil, ErrKeyNotFound
	}
	return AllIDUrls, nil
}

func (fs *FileStorage) Close() {
	fs.FileJSON.Close()
}

func (fs *FileStorage) BatchURL(_ctx context.Context, userID string, urls []Correlation) ([]Correlation, error) {
	// var shortID string
	// urlslen := len(urls)
	// newurls := make([]Correlation, urlslen)
	// for i := range urls {
	// 	shortID = GenerateRandomString()
	// 	shorturl := BaseTest + shortID
	// 	newurls[i].ShortURL = shorturl
	// 	newurls[i].UserID = userID
	// 	newurls[i].OriginalURL = urls[i].OriginalURL
	// 	newurls[i].CorrelationID = urls[i].CorrelationID
	// }
	if fs.NewMap == nil {
		fs.NewMap = make(map[string]map[string]string)
	}
	if fs.NewMap[userID] == nil {
		fs.NewMap[userID] = make(map[string]string)
	}
	if err := os.Truncate("OurURL.json", 0); err != nil {
		return nil, fmt.Errorf("BatchURL: unable to Truncate file: %w", err)
	}
	// for i := range newurls {
	// 	if newurls[i].ShortURL == "" {
	// 		return nil, ErrKeyNotSpecified
	// 	}
	// 	fs.NewMap[newurls[i].UserID][newurls[i].ShortURL] = newurls[i].OriginalURL
	// }
	var newurls []Correlation
	for _, batch := range urls {
		shortID := GenerateRandomString()
		shorturl := BaseTest + shortID
		batch.ShortURL = shorturl
		newurl := Correlation{
			ShortURL:      batch.ShortURL,
			UserID:        userID,
			OriginalURL:   batch.OriginalURL,
			CorrelationID: batch.CorrelationID,
		}
		newurls = append(newurls, newurl)
		fs.NewMap[batch.UserID][batch.ShortURL] = batch.OriginalURL
	}
	jsonData, err := json.Marshal(fs.NewMap)
	if err != nil {
		return nil, fmt.Errorf("BatchURL: unable to marshal internal file storage map: %w", err)
	}
	fs.FileJSON.Write(jsonData)
	return newurls, nil
}
