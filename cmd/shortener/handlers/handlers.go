package handlers

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/configuration"
	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/service/repository"
	"github.com/go-chi/chi/v5"
)

const (
	symbols        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	ShortURLMaxLen = 7
)

func generateRandomString() string {
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, 0, ShortURLMaxLen)
	for i := 0; i < ShortURLMaxLen; i++ {
		s := symbols[rand.Intn(len(symbols))]
		result = append(result, s)
	}
	return string(result)
}

func PostHandler(repo repository.AddorGetURL, Cfg configuration.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortID := generateRandomString()
		// Здесь мы проверяем что урл который мы сгенерировали отсутствует в базе
		// если же он там есть, мы перегенерируем и так пока не получим уникальный
		for _, err := repo.GetURL(shortID); !errors.Is(err, repository.ErrKeyNotFound); _, err = repo.GetURL(shortID) {
			if err != nil {
				log.Printf("unable to get URL by short ID: %v", err)
				http.Error(w, "unable to get url from DB", http.StatusInternalServerError)
				return
			}
			shortID = generateRandomString()
		}
		shorturl := configuration.Cfg.ConfigURL.JoinPath(shortID)
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
		ourPoorURL := repository.URL{ShortURL: shortID, OriginalURL: longURL}
		err = repo.AddURL(ourPoorURL)
		if err != nil {
			http.Error(w, "Status internal server error", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shorturl.String()))
	}
}

func GetHandler(repo repository.AddorGetURL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortnew := chi.URLParam(r, "id")
		originalURL, err := repo.GetURL(shortnew)
		if err != nil {
			http.Error(w, "unable to GET Original url", http.StatusBadRequest)
			return
		}
		if err != nil {
			log.Printf("unable to get key from repo: %v", err)
			http.Error(w, "unable to GET Original url", http.StatusInternalServerError)
			return
		}
		if errors.Is(err, repository.ErrKeyNotExists) {
			http.Error(w, "unable to GET Original url", http.StatusBadRequest)
			return
		}
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func PostJSONHandler(repo repository.AddorGetURL, Cfg configuration.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tempStrorage repository.JSONKeymap
		if err := json.NewDecoder(r.Body).Decode(&tempStrorage); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		longURL, err := url.QueryUnescape(tempStrorage.LongJSON)
		if err != nil {
			http.Error(w, "unable to QueryUnescape longURL", http.StatusBadRequest)
			return
		}
		shortID := generateRandomString()
		for _, err := repo.GetURL(shortID); !errors.Is(err, repository.ErrKeyNotFound); _, err = repo.GetURL(shortID) {
			if err != nil {
				log.Printf("unable to get URL by short ID: %v", err)
				http.Error(w, "unable to get url from DB", http.StatusInternalServerError)
				return
			}
			shortID = generateRandomString()
		}
		ourPoorURL := repository.URL{ShortURL: shortID, OriginalURL: longURL}
		err = repo.AddURL(ourPoorURL)
		if err != nil {
			http.Error(w, "Status internal server error", http.StatusBadRequest)
			return
		}
		shortURL := configuration.Cfg.ConfigURL.JoinPath(shortID)
		resobj := repository.JSONKeymap{ShortJSON: shortURL.String(), LongJSON: longURL}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(&resobj)
	}
}

func DecomprMiddlw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(gz)
			defer gz.Close()
		}
		next.ServeHTTP(w, r)
	})
}
