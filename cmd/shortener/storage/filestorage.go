package storage

import (
	"encoding/json"
	"os"
	// "os"
)

type JSONKeymap struct {
	ShortJSON string `json:"result,omitempty"`
	LongJSON  string `json:"url,omitempty"`
}

// type FileStorage struct {
// 	allurls map[string]string
// 	fileJSON os.File
// }

// type AddGetFileInterf interface {
// 	AddURL(fileadress string, longandshortURL repository.URL) error
// 	GetURL(idshortURL string) (string, error)
// }

// func NewFileStorage () AddGetFileInterf {
// 	return &FileStorage {
// 		allurls: make(map[string]string),
// 		// file: os.OpenFile(),
// 	}
// }

// func (fs *FileStorage) AddURL(f string, u repository.URL) error {
// 	if fs.allurls == nil {
// 		fs.allurls = make(map[string]string)
// 	}
// 	if u.ShortURL == "" {
// 		return repository.ErrorKeyNotSpecified
// 	}
// 	fs.allurls[u.ShortURL] = u.OriginalURL
// 	file, err := os.OpenFile(f, os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fs.fileJSON = *file
// 	jsonData, err := json.Marshal(fs.allurls)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fs.fileJSON.Write(jsonData)
// 	defer fs.fileJSON.Close()
// 	return nil
// }

// func (fs *FileStorage) GetURL(key string) (string, error) {
// 	if key == "" {
// 		return "", repository.ErrorKeyNotSpecified
// 	}
// 	if value, ok := fs.allurls[key]; ok {
// 		return value, nil
// 	}
// 	return "", repository.ErrorKeyNotFound
// }

func FileWriteFunc(fileadress, short, longURL string) {
	file, err := os.OpenFile(fileadress, os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		panic(err)
	}
	urlmap := make(map[string]string)
	urlmap[short] = longURL
	jsonData, err := json.Marshal(urlmap)
	if err != nil {
		panic(err)
	}
	file.Write(jsonData)
	defer file.Close()
}

func FileReadFunc(fileadress string) (resultshort map[string]string) {
	file, err := os.ReadFile(fileadress)
	if err != nil {
		return nil
	}
	var byteData map[string]string
	if err := json.Unmarshal(file, &byteData); err != nil {
		return nil
	}
	return byteData
}
