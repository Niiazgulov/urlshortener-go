package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

// func RemoveChar(word string) string {
// 	return word[1:]
// }

// func NewRouter() chi.Router {
// 	r := chi.NewRouter()
// 	r.Use(middleware.RequestID)
// 	r.Use(middleware.RealIP)
// 	r.Use(middleware.Logger)
// 	r.Use(middleware.Recoverer)
// 	r.Route("/", func(r chi.Router) {
// 		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
// 			rand.Seed(time.Now().UnixNano())
// 			randint := rand.Uint64()
// 			short := Encoder(randint)
// 			shorturl := "http://localhost:8080/" + short
// 			longURLByte, err := io.ReadAll(r.Body)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			longURL := strings.ReplaceAll(string(longURLByte), "url=", "")
// 			longURL, _ = url.QueryUnescape(longURL)
// 			keymap[short] = longURL
// 			w.WriteHeader(http.StatusCreated)
// 			w.Write([]byte(shorturl))
// 		})
// 		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
// 			shortnew := chi.URLParam(r, "id")
// 			originalURL := keymap[shortnew]
// 			w.Header().Set("Location", originalURL)
// 			w.WriteHeader(http.StatusTemporaryRedirect)
// 		})
// 	})

// 	return r
// }

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Route("/", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
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
		})
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			shortnew := chi.URLParam(r, "id")
			originalURL := keymap[shortnew]
			w.Header().Set("Location", originalURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
		})
	})
	log.Fatal(http.ListenAndServe(":8080", r))
}
