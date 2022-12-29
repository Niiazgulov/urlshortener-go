package main

import (
	// "io/ioutil"
	// "log"
	// "fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	symbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// var (
//     keymap = make(map[string]string)
// )

func Encoder(number uint64) string {
	length := len(symbols)
	var encodedBuilder strings.Builder
	encodedBuilder.Grow(10)
	for ; number > 0; number = number / uint64(length) {
		encodedBuilder.WriteByte(symbols[(number % uint64(length))])
	}

	return encodedBuilder.String()
}

// Обработчик запроса
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Only POST or GET requests are allowed!", http.StatusBadRequest)
		return
	}

	longurl := r.URL.Path
	if longurl == "" {
		http.Error(w, "This URL is empty", http.StatusBadRequest)
		return
	}
	rand.Seed(time.Now().UnixNano())
	randint := rand.Uint64()
	id := Encoder(randint)
	shorturl := "http://localhost:8080/" + id
	keymap := make(map[string]string)
	keymap[id] = longurl

	switch r.Method {
	case http.MethodPost:
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shorturl))
	case http.MethodGet:
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte(longurl))
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "Bad Request"}`))
	}

}

func main() {
	// server := &http.Server{
	// 	Addr: "mydomain.com:80",
	// }

	// маршрутизация запросов обработчику
	http.HandleFunc("POST /", Handler)
	http.HandleFunc("GET /{id}", Handler)
	// // запуск сервера с адресом localhost, порт 8080
	http.ListenAndServe(":8080", nil)

	// log.Fatal(server.ListenAndServe())
}
