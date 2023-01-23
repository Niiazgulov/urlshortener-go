package handlers

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/configuration"
	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/service/repository"
	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/storage"
	"github.com/go-chi/chi/v5"
)

var (
	repo repository.AddorGetURL
)

func init() {
	repo = repository.NewMemoryRepository()
}

const (
	symbols        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	BaseURL        = "http://localhost:8080/"
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

func PostHandler(w http.ResponseWriter, r *http.Request) {
	shortID := generateRandomString()
	for _, err := repo.GetURL(shortID); err == nil; _, err = repo.GetURL(shortID) {
		shortID = generateRandomString()
	}
	shortParse, err := url.Parse(shortID)
	if err != nil {
		http.Error(w, "unable to Parse shortID", http.StatusBadRequest)
		return
	}
	short := shortParse.String()
	BaseCfgURL, _ := url.Parse(configuration.Cfg.BaseURLAddress)
	shorturl := BaseCfgURL.JoinPath(short)
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
	ourPoorURL := repository.URL{ShortURL: short, OriginalURL: longURL}
	err = repo.AddURL(ourPoorURL)
	if err != nil {
		http.Error(w, "Status internal server error", http.StatusBadRequest)
		return
	}
	if configuration.Cfg.FilePath != "" {
		storage.FileWriteFunc(configuration.Cfg.FilePath, short, longURL)
	} else {
		err = repo.AddURL(ourPoorURL)
		if err != nil {
			http.Error(w, "Status internal server error", http.StatusBadRequest)
			return
		}
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shorturl.String()))
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	shortnew := chi.URLParam(r, "id")
	// var originalURL string
	// if configuration.Cfg.FilePath != "" {
	// 	byteInfo := storage.FileReadFunc(configuration.Cfg.FilePath)
	// 	originalURL = byteInfo[shortnew]
	// } else {
	// 	var err error
	// 	originalURL, err = repo.GetURL(shortnew)
	// 	if err != nil {
	// 		http.Error(w, "unable to GET Original url", http.StatusBadRequest)
	// 		return
	// 	}
	// }
	originalURL, err := repo.GetURL(shortnew)
	if err != nil {
		http.Error(w, "unable to GET Original url", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func PostJSONHandler(w http.ResponseWriter, r *http.Request) {
	var tempStrorage storage.JSONKeymap
	if err := json.NewDecoder(r.Body).Decode(&tempStrorage); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	longURL, err := url.QueryUnescape(tempStrorage.LongJSON)
	if err != nil {
		http.Error(w, "unable to QueryUnescape longURL", http.StatusBadRequest)
		return
	}
	shortID, err := url.Parse(generateRandomString())
	if err != nil {
		http.Error(w, "unable to Parse shortID", http.StatusBadRequest)
		return
	}
	ourPoorURL := repository.URL{ShortURL: shortID.String(), OriginalURL: longURL}
	err = repo.AddURL(ourPoorURL)
	if err != nil {
		http.Error(w, "Status internal server error", http.StatusBadRequest)
		return
	}
	BaseCfgURL, _ := url.Parse(configuration.Cfg.BaseURLAddress)
	shortURL := BaseCfgURL.JoinPath(shortID.String())
	resobj := storage.JSONKeymap{ShortJSON: shortURL.String(), LongJSON: longURL}
	if configuration.Cfg.FilePath != "" {
		storage.FileWriteFunc(configuration.Cfg.FilePath, shortURL.String(), longURL)
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&resobj)
}

// type gzipWriter struct {
// 	http.ResponseWriter
// 	Writer io.Writer
// }

// func ComprMiddlw(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
// 			next.ServeHTTP(w, r)
// 			return
// 		}
// 		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
// 		if err != nil {
// 			io.WriteString(w, err.Error())
// 			return
// 		}
// 		defer gz.Close()
// 		w.Header().Set("Content-Encoding", "gzip")
// 		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
// 	})
// }

func UncomprMiddlw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reader io.Reader

		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			reader = gz
			r.Body = io.NopCloser(reader)
			defer gz.Close()
		} else {
			reader = r.Body
		}

		next.ServeHTTP(w, r)
	})
}
