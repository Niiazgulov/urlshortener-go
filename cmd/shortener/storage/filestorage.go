package storage

import (
	"encoding/json"
	"log"
	"os"
	// "github.com/Niiazgulov/urlshortener.git/cmd/shortener/service/repository"
	//"github.com/Niiazgulov/urlshortener.git/cmd/shortener/configuration"
)

type JSONKeymap struct {
	ShortJSON string `json:"result,omitempty"`
	LongJSON  string `json:"url,omitempty"`
}

type saver struct {
	file    *os.File
	encoder *json.Encoder
}

func NewSaver(fileName string) (*saver, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	return &saver{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}
func (s *saver) WriteKeymap(keymap *JSONKeymap) error {
	return s.encoder.Encode(&keymap)
}

//	func (s *saver) WriteJSONKeymap(keymap *JSONKeymap) error {
//		return s.encoder.Encode(&keymap)
//	}
func (s *saver) Close() error {
	return s.file.Close()
}

func FileWriteFunc(fileadress, short, longURL string) {
	file, err := os.OpenFile(fileadress, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		log.Fatal(err)
	}
	urlmap := make(map[string]string)
	urlmap[short] = longURL
	jsonData, err := json.Marshal(urlmap)
	if err != nil {
		log.Fatal(err)
	}
	file.Write(jsonData)
	defer file.Close()
}

type loader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewLoader(fileName string) (*loader, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &loader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}
func (l *loader) ReadKeymap() (*JSONKeymap, error) {
	kmp := &JSONKeymap{}
	if err := l.decoder.Decode(&kmp); err != nil {
		return nil, err
	}
	return kmp, nil
}
func (l *loader) Close() error {
	return l.file.Close()
}
