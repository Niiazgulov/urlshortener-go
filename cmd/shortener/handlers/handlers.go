package handlers

import (
	"compress/gzip"
	"encoding/json"
	"io"

	//"os"
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

// var (
// 	// repo repository.AddGetFileInterf
// 	repo repository.AddorGetURL
// 	// repofile storage.AddGetFileInterf
// )

// func init() {
// 	repo = repository.NewFileStorage()
// 	//repo = repository.NewMemoryRepository()
// 	// repofile = storage.NewFileStorage()
// 	// var err error
// 	// repository.FileTemp, err = os.OpenFile(configuration.Cfg.FilePath, os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// 	// defer repository.FileTemp.Close()
// }

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
		for _, err := repo.GetURL(shortID); err == nil; _, err = repo.GetURL(shortID) {
			shortID = generateRandomString()
		}
		shortParse, err := url.Parse(shortID)
		if err != nil {
			http.Error(w, "unable to Parse shortID", http.StatusBadRequest)
			return
		}
		short := shortParse.String()
		shorturl := configuration.Cfg.ConfigUrl.JoinPath(short)
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
		// if configuration.Cfg.FilePath != "" {
		// 	storage.FileWriteFunc(configuration.Cfg.FilePath, short, longURL)
		// }
		// if configuration.Cfg.FilePath != "" {
		// 	storage.FileWriteFunc(configuration.Cfg.FilePath, short, longURL)
		// } else {
		// 	err = repo.AddURL(ourPoorURL)
		// 	if err != nil {
		// 		http.Error(w, "Status internal server error", http.StatusBadRequest)
		// 		return
		// 	}
		// }
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
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func PostJSONHandler(repo repository.AddorGetURL, Cfg configuration.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		shortURL := configuration.Cfg.ConfigUrl.JoinPath(shortID.String())
		resobj := storage.JSONKeymap{ShortJSON: shortURL.String(), LongJSON: longURL}
		// if configuration.Cfg.FilePath != "" {
		// 	storage.FileWriteFunc(configuration.Cfg.FilePath, shortURL.String(), longURL)
		// }
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(&resobj)
	}
}

func DecomprMiddlw(next http.Handler) http.Handler {
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
