package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	keymap = make(map[string]string, 100)
)

const (
	symbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func Encoder(number uint64) string {
	length := len(symbols)
	var encodedBuilder strings.Builder
	encodedBuilder.Grow(10)
	for ; number > 0; number = number / uint64(length) {
		encodedBuilder.WriteByte(symbols[(number % uint64(length))])
	}
	return encodedBuilder.String()
}

func RemoveChar(word string) string {
	return word[1:]
}

func BestHandlerEver(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Only GET or POST requests are allowed!", http.StatusBadRequest)
		return
	}
	longURL := r.URL.Path
	if longURL == "" {
		http.Error(w, "This URL is empty", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPost:
		rand.Seed(time.Now().UnixNano())
		randint := rand.Uint64()
		short := Encoder(randint)
		shorturl := "http://localhost:8080/" + short
		keymap[short] = longURL
		jsonData, err := json.Marshal(keymap)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(jsonData))
		jsonFile, err := os.Create("./OurURL.json")
		if err != nil {
			panic(err)
		}
		defer jsonFile.Close()
		jsonFile.Write(jsonData)
		jsonFile.Close()
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shorturl))
	case http.MethodGet:
		short := r.URL.Path
		var data []byte
		data, _ = ioutil.ReadFile("OurURL.json")
		var m map[string]string
		err := json.Unmarshal(data, &m)
		if err != nil {
			log.Fatal(err)
		}
		shortnew := RemoveChar(short)
		originalURL := m[shortnew]
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		short2 := r.URL.Path
		originalURL2 := keymap[short2]
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Location", originalURL2)
	}
}

func main() {
	http.HandleFunc("/", BestHandlerEver)
	http.ListenAndServe(":8080", nil)
}
