package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

type FileStorage struct {
	FileJSON *os.File
	// Ключ мапы ShortID
	urlMap map[string]URL
	mutex  sync.RWMutex
}

func NewFileStorage(f *os.File) (AddorGetURL, error) {
	m := make(map[string]URL)
	err := json.NewDecoder(f).Decode(&m)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("unable to unmarshal file into map: %w", err)
	}
	if errors.Is(err, io.EOF) {
		m = make(map[string]URL)
	}
	return &FileStorage{
		FileJSON: f,
		urlMap:   m,
		mutex:    sync.RWMutex{},
	}, nil
}

func (fs *FileStorage) AddURL(u URL) error {
	if fs.urlMap == nil {
		fs.urlMap = make(map[string]URL)
	}
	if u.ShortURL == "" {
		return ErrKeyNotSpecified
	}
	u.Deleted = false
	fs.urlMap[u.ShortURL] = u
	jsonData, err := json.Marshal(fs.urlMap)
	if err != nil {
		return fmt.Errorf("unable to marshal internal file storage map: %w", err)
	}
	if err := os.Truncate("OurURL.json", 0); err != nil {
		return fmt.Errorf("unable to Truncate file: %w", err)
	}
	fs.FileJSON.Write(jsonData)
	return nil
}

func (fs *FileStorage) GetOriginalURL(_ context.Context, shortID string) (string, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	if shortID == "" {
		return "", ErrKeyNotSpecified
	}
	for _, v := range fs.urlMap {
		if shortID == v.ShortURL {
			if v.Deleted {
				return "", ErrURLdeleted
			}
			return v.OriginalURL, nil
		}
	}
	return "", ErrKeyNotFound
}

func (fs *FileStorage) GetShortURL(_ context.Context, originalurl string) (string, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	if originalurl == "" {
		return "", ErrKeyNotSpecified
	}
	for _, v := range fs.urlMap {
		if v.OriginalURL == originalurl {
			return v.ShortURL, nil
		}
	}
	return "", ErrKeyNotFound
}

func (fs *FileStorage) FindAllUserUrls(_ context.Context, userID string) (map[string]string, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	m := make(map[string]string)
	for _, v := range fs.urlMap {
		if v.UserID == userID {
			m[v.ShortURL] = v.OriginalURL
		}
	}
	return m, nil
}

func (fs *FileStorage) BatchURL(_ctx context.Context, userID string, urls []URL) ([]ShortCorrelation, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	if fs.urlMap == nil {
		fs.urlMap = make(map[string]URL)
	}
	err := fs.FileJSON.Truncate(0)
	if err != nil {
		return nil, fmt.Errorf("batchurl: unable to truncate file: %w", err)
	}
	_, err = fs.FileJSON.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("batchurl: unable to get the beginning of file: %w", err)
	}
	var newurls []ShortCorrelation
	var batchurl URL
	for _, batch := range urls {
		shortID := GenerateRandomString()
		shorturl := BaseTest + shortID
		newurl := ShortCorrelation{
			ShortURL:      shorturl,
			CorrelationID: batch.CorrelationID,
		}
		newurls = append(newurls, newurl)
		batchurl = URL{ShortURL: shortID, OriginalURL: batch.OriginalURL}
		fs.urlMap[batch.ShortURL] = batchurl
	}
	jsonData, err := json.Marshal(fs.urlMap)
	if err != nil {
		return nil, fmt.Errorf("BatchURL: unable to marshal internal file storage map: %w", err)
	}
	fs.FileJSON.Write(jsonData)
	return newurls, nil
}

func (fs *FileStorage) DeleteUrls(urls []URL) error {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	for _, urlforDelete := range urls {
		markURL, ok := fs.urlMap[urlforDelete.ShortURL]
		if ok && markURL.UserID == urlforDelete.UserID {
			markURL.Deleted = true
			fs.urlMap[urlforDelete.ShortURL] = markURL
		}
	}
	jsonData, err := json.Marshal(fs.urlMap)
	if err != nil {
		return fmt.Errorf("deleteurls: unable to marshal internal file storage map: %w", err)
	}
	if err := os.Truncate("OurURL.json", 0); err != nil {
		return fmt.Errorf("deleteurls: unable to Truncate file: %w", err)
	}
	// fs.FileJSON.Truncate(0)
	fs.FileJSON.Write(jsonData)
	return nil
}

func (fs *FileStorage) Close() {
	fs.FileJSON.Close()
}
