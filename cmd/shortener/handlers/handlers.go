package handlers

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/service/repository"
	// "github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
)

var repo repository.AddorGetURL

func init() {
	repo = repository.NewMemoryRepository()
}

const (
	symbols        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	BaseURL        = "http://localhost:8080/"
	ShortURLMaxLen = 7
)

type JSONKeymap struct {
	ShortJSON string `json:"result,omitempty"`
	LongJSON  string `json:"url,omitempty"`
}

type Config struct {
	ServerAddress  string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURLAddress string `env:"BASE_URL" envDefault:"http://localhost:8080/"`
}

// var (
// 	Cfg = Config{BaseURLAddress: "http://localhost:8080/", ServerAddress: ":8080"}
// )

func generateRandomString() string {
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, 0, ShortURLMaxLen)
	for i := 0; i < ShortURLMaxLen; i++ {
		s := symbols[rand.Intn(len(symbols))]
		result = append(result, s)
	}
	return string(result)
}

func PostURLHandler(w http.ResponseWriter, r *http.Request) {
	short := generateRandomString()
	for _, err := repo.GetURL(short); err == nil; _, err = repo.GetURL(short) {
		short = generateRandomString()
	}
	// shorturl := Cfg.BaseURLAddress + short
	shorturl := BaseURL + short
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
	// Cfg.BaseURLAddress = shorturl
	// Cfg.BaseURLAddress = longURL
	// err = env.Parse(&Cfg)
	// if err != nil {
	// 	http.Error(w, "Can't Parse Config (env)", http.StatusBadRequest)
	// 	return
	// }
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shorturl))
}

func GetURLHandler(w http.ResponseWriter, r *http.Request) {
	shortnew := chi.URLParam(r, "id")
	originalURL, err := repo.GetURL(shortnew)
	if err != nil {
		http.Error(w, "unable to GET Original url", http.StatusBadRequest)
		return
	}
	// var originalURL string
	// if Cfg.BaseURLAddress != "" {
	// 	originalURL = Cfg.BaseURLAddress
	// } else {
	// 	var err error
	// 	originalURL, err = repo.GetURL(shortnew)
	// 	if err != nil {
	// 		http.Error(w, "unable to GET Original url", http.StatusBadRequest)
	// 		return
	// 	}
	// }
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func PostJSONHandler(w http.ResponseWriter, r *http.Request) {
	longURLByte, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read Body", http.StatusBadRequest)
		return
	}
	JSONrequest := JSONKeymap{}
	if err := json.Unmarshal([]byte(longURLByte), &JSONrequest); err != nil {
		http.Error(w, "unable to unescape JSON query in input url", http.StatusBadRequest)
		return
	}
	longURL := JSONrequest.LongJSON
	short := generateRandomString()
	for _, err := repo.GetURL(short); err == nil; _, err = repo.GetURL(short) {
		short = generateRandomString()
	}
	// shorturl := Cfg.BaseURLAddress + short
	shorturl := BaseURL + short
	// Cfg.BaseURLAddress = longURL
	//Cfg.BaseURLAddress = shorturl
	// err = env.Parse(&Cfg)
	// if err != nil {
	// 	http.Error(w, "Can't Parse Config (env)", http.StatusBadRequest)
	// 	return
	// }
	JSONresponse := JSONKeymap{ShortJSON: shorturl}
	response, err := json.Marshal(JSONresponse)
	if err != nil {
		http.Error(w, "Can't make a json.Marshal operation", http.StatusBadRequest)
		return
	}
	ourPoorURL := repository.URL{ShortURL: short, OriginalURL: longURL}
	err = repo.AddURL(ourPoorURL)
	if err != nil {
		http.Error(w, "Status internal server error", http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
