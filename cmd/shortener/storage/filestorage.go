package storage

import (
	"encoding/json"
	// "io/ioutil"
	"log"
	"os"
)

type JSONKeymap struct {
	ShortJSON string `json:"result,omitempty"`
	LongJSON  string `json:"url,omitempty"`
}

func FileWriteFunc(fileadress, short, longURL string) {
	file, err := os.OpenFile(fileadress, os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
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

func FileReadFunc(fileadress string) (resultshort map[string]string) {
	// file, err := ioutil.ReadFile(fileadress)
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
