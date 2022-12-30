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

func BestHandlerEver(w http.ResponseWriter, r *http.Request) {
	// этот обработчик принимает только запросы, отправленные методом POST
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Only GET or POST requests are allowed!", http.StatusBadRequest)
		return
	}
	//URLid := r.URL.Query().Get("id")
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
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shorturl))
	case http.MethodGet:
		short := r.URL.RequestURI()
		originalURL := keymap[short]
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Header().Set("Location", originalURL)
	default:
		short2 := r.URL.Path
		originalURL2 := keymap[short2]
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Location", originalURL2)
	}
}

// func GetHandler(w http.ResponseWriter, r *http.Request) {
// 	// этот обработчик принимает только запросы, отправленные методом GET
// 	if r.Method != http.MethodGet {
// 		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
// 		return
// 	}
// 	short
// 	//URLid := r.URL.Query().Get("/")
// 	URLid := r.URL.Path
// 	originalURL := keymap[URLid]
// 	w.Header().Set("Location", originalURL)
// 	w.WriteHeader(http.StatusTemporaryRedirect)
// }

func main() {
	// маршрутизация запросов обработчику
	http.HandleFunc("/", BestHandlerEver)
	// http.HandleFunc("/{keymap[shorturl]}", GetHandler)
	// // запуск сервера с адресом localhost, порт 8080
	http.ListenAndServe(":8080", nil)
	// log.Fatal(server.ListenAndServe())
}
