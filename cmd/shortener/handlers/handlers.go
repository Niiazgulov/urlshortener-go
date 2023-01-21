package handlers

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
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

type JSONKeymap struct {
	ShortJSON string `json:"result,omitempty"`
	LongJSON  string `json:"url,omitempty"`
}

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
	// file, err := os.OpenFile("./OurURL.json", os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	// if err != nil {
	// 	http.Error(w, "error while opening the file", http.StatusBadRequest)
	// 	return
	// }
	// urlmap := make(map[string]string)
	// urlmap[short] = longURL
	// jsonData, err := json.Marshal(urlmap)
	// if err != nil {
	// 	http.Error(w, "unable to Marshal urlmap", http.StatusBadRequest)
	// 	return
	// }
	// file.Write(jsonData)
	// defer file.Close()
	//filename := "events.log"
	configuration.Cfg.FilePath = "OurURL.json"
	storage.FileWriteFunc(configuration.Cfg.FilePath, short, longURL)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shorturl.String()))
}

func GetURLHandler(w http.ResponseWriter, r *http.Request) {
	shortnew := chi.URLParam(r, "id")

	var originalURL string
	if configuration.Cfg.FilePath != "" {
		// file, err := os.OpenFile("./OurURL.json", os.O_RDONLY, 0777)
		// if err != nil {
		// 	http.Error(w, "Error while opening the file", http.StatusBadRequest)
		// 	return
		// }
		newkeymap := make(map[string]string)
		fileBytes, _ := os.ReadFile("./OurURL.json")
		err := json.Unmarshal(fileBytes, &newkeymap)
		if err != nil {
			http.Error(w, "Error while opening the file", http.StatusBadRequest)
			return
		}
		originalURL = newkeymap[shortnew]
	} else {
		var err error
		originalURL, err = repo.GetURL(shortnew)
		if err != nil {
			http.Error(w, "unable to GET Original url", http.StatusBadRequest)
			return
		}
	}
	// originalURL, err := repo.GetURL(shortnew)
	// if err != nil {
	// 	http.Error(w, "unable to GET Original url", http.StatusBadRequest)
	// 	return
	// }

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
	var tempStrorage JSONKeymap
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
	resobj := JSONKeymap{ShortJSON: shortURL.String()}
	configuration.Cfg.FilePath = "OurURL.json"
	storage.FileWriteFunc(configuration.Cfg.FilePath, shortID.String(), longURL)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&resobj)
}
