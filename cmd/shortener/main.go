package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	symbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var keymap = map[string]string{}

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
	switch r.Method {
	case http.MethodPost:
		rand.Seed(time.Now().UnixNano())
		randint := rand.Uint64()
		short := Encoder(randint)
		shorturl := "http://localhost:8080/" + short
		longURLByte, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		longURL := strings.ReplaceAll(string(longURLByte), "url=", "")
		longURL, _ = url.QueryUnescape(longURL)
		keymap[short] = longURL
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shorturl))
	case http.MethodGet:
		short := r.URL.Path
		shortnew := RemoveChar(short)
		originalURL := keymap[shortnew]
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		short2 := r.URL.Path
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Location", short2)
	}
}

func main() {
	http.HandleFunc("/", BestHandlerEver)
	http.ListenAndServe(":8080", nil)
}
