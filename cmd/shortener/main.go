package main

import (
	// "fmt"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
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
	var data []byte
	var mapu map[string]string
	data, _ = ioutil.ReadFile("OurURL.json")
	err := json.Unmarshal(data, &mapu)
	if err != nil {
		log.Fatal(err)
	}

	switch r.Method {
	case http.MethodPost:
		rand.Seed(time.Now().UnixNano())
		randint := rand.Uint64()
		short := Encoder(randint)
		shorturl := "http://localhost:8080/" + short
		longURL := r.URL.Path
		if longURL == "" {
			http.Error(w, "This URL is empty", http.StatusBadRequest)
			return
		}
		mapu[short] = longURL
		jsonData, err := json.Marshal(mapu)
		if err != nil {
			panic(err)
		}
		file, err := os.OpenFile("./OurURL.json", os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			log.Fatalf("error while opening the file. %v", err)
		}
		file.Write(jsonData)
		file.Close()
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shorturl))
	case http.MethodGet:
		short := r.URL.Path
		shortnew := RemoveChar(short)
		originalURL := mapu[shortnew]
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		short2 := r.URL.Path
		//originalURL2 := keymap[short2]
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Location", short2)
	}
}

func main() {
	http.HandleFunc("/", BestHandlerEver)
	http.ListenAndServe(":8080", nil)
}
