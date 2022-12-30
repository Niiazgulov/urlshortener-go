package main

import (
	"math/rand"
	"net/http"
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

func PostHandler(w http.ResponseWriter, r *http.Request) {
	// этот обработчик принимает только запросы, отправленные методом POST
	if r.Method != http.MethodPost {
		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}
	//URLid := r.URL.Query().Get("id")
	originalURL := r.URL.Path
	if originalURL == "" {
		http.Error(w, "This URL is empty", http.StatusBadRequest)
		return
	}
	rand.Seed(time.Now().UnixNano())
	randint := rand.Uint64()
	short := Encoder(randint)
	shorturl := "http://localhost:8080/" + short
	keymap[shorturl] = originalURL
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shorturl))
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	// этот обработчик принимает только запросы, отправленные методом GET
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}
	//URLid := r.URL.Query().Get("/")
	URLid := r.URL.Path
	originalURL := keymap[URLid]
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func main() {
	// маршрутизация запросов обработчику
	http.HandleFunc("/", PostHandler)
	http.HandleFunc("/{keymap[shorturl]}", GetHandler)
	// // запуск сервера с адресом localhost, порт 8080
	http.ListenAndServe(":8080", nil)
	// log.Fatal(server.ListenAndServe())
}
