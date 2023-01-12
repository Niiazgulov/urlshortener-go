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
	baseURL = "http://localhost:8080/"
)

func Encoder() string {
	result := make([]byte, 0, 7)
	for i := 0; i < 7; i++ {
		s := symbols[rand.Intn(len(symbols))]
		result = append(result, s)
	}
	return string(result)
}

func NewRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", GetURLHandler)
		r.Post("/", PostURLHandler)
	})

	return r
}

func main() {
	rand.Seed(time.Now().UnixNano())
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Route("/", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			short := Encoder()
			shorturl := baseURL + short
			longURLByte, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "can't read Body", http.StatusBadRequest)
				return
			}
			longURL := strings.TrimPrefix(string(longURLByte), "url=")
			longURL, err = url.QueryUnescape(longURL)
			if err != nil {
				http.Error(w, "unable to unescape query in input url", http.StatusBadRequest)
				return
			}
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
