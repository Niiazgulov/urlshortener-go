package handlers

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/url"

	// "os"
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
	if configuration.Cfg.FilePath != "" {
		storage.FileWriteFunc(configuration.Cfg.FilePath, short, longURL)
	} else {
		err = repo.AddURL(ourPoorURL)
		if err != nil {
			http.Error(w, "Status internal server error", http.StatusBadRequest)
			return
		}
	}

	// vers 3
	// fileName := configuration.Cfg.FilePath
	// // defer os.Remove(fileName)
	// producer, err := storage.NewProducer(fileName)
	// if err != nil {
	// 	http.Error(w, "Can't create file", http.StatusBadRequest)
	// 	// 	return
	// }
	// defer producer.Close()
	// events := storage.Event{ShortJSON: short, LongJSON: longURL}
	// if err := producer.WriteEvent(&events); err != nil {
	// 	http.Error(w, "Can't WriteEvent()", http.StatusBadRequest)
	// 	return
	// }
	// vers 3.1
	// for _, event := range events {
	// 	if err := producer.WriteEvent(event); err != nil {
	// 		http.Error(w, "Can't WriteEvent()", http.StatusBadRequest)
	// 		return
	// 	}
	// }
	// vers 2
	// configuration.Cfg.FilePath = "OurURL.json"
	// fileName := configuration.Cfg.FilePath
	// defer os.Remove(fileName)
	// saver, err := storage.NewSaver(fileName)
	// if err != nil {
	// 	http.Error(w, "Can't create file", http.StatusBadRequest)
	// 	return
	// }
	// defer saver.Close()
	// resobj := storage.JSONKeymap{ShortJSON: short, LongJSON: longURL}
	// if err := saver.WriteKeymap(&resobj); err != nil {
	// 	http.Error(w, "Can't save info to the file", http.StatusBadRequest)
	// 	return
	// }
	// vers 1
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shorturl.String()))
}

func GetURLHandler(w http.ResponseWriter, r *http.Request) {
	shortnew := chi.URLParam(r, "id")
	var originalURL string
	if configuration.Cfg.FilePath != "" {
		byteInfo := storage.FileReadFunc(configuration.Cfg.FilePath)
		originalURL = byteInfo[shortnew]
	} else {
		var err error
		originalURL, err = repo.GetURL(shortnew)
		if err != nil {
			http.Error(w, "unable to GET Original url", http.StatusBadRequest)
			return
		}
	}
	// vers Last2
	// fileName := configuration.Cfg.FilePath
	// consumer, err := storage.NewConsumer(fileName)
	// if err != nil {
	// 	http.Error(w, "Can't create NewConsumer()", http.StatusBadRequest)
	// 	return
	// }
	// defer consumer.Close()
	// readEvent, err := consumer.ReadEvent()
	// if err != nil {
	// 	http.Error(w, "Can't ReadEvent()", http.StatusBadRequest)
	// 	return
	// }
	// var originalURL string
	// if readEvent.ShortJSON == shortnew {
	// 	originalURL = readEvent.LongJSON
	// } else {
	// 	originalURL, err = repo.GetURL(shortnew)
	// 	if err != nil {
	// 		http.Error(w, "unable to GET Original url", http.StatusBadRequest)
	// 		return
	// 	}
	// }

	// vers Last
	// var originalURL string
	// configuration.Cfg.FilePath = "OurURL.json"
	// fileName := configuration.Cfg.FilePath
	// loader, err := storage.NewLoader(fileName)
	// if err != nil {
	// 	http.Error(w, "Can't load info from the file", http.StatusBadRequest)
	// 	return
	// }
	// defer loader.Close()
	// readEvent, err := loader.ReadKeymap()
	// if err != nil {
	// 	http.Error(w, "Error while ReadKeymap()", http.StatusBadRequest)
	// 	return
	// }
	// if readEvent.ShortJSON == shortnew {
	// 	originalURL = readEvent.LongJSON
	// } else {
	// 	originalURL, err = repo.GetURL(shortnew)
	// 	if err != nil {
	// 		http.Error(w, "unable to GET Original url", http.StatusBadRequest)
	// 		return
	// 	}
	// }

	// vers X3
	// var originalURL string
	// if configuration.Cfg.FilePath != "" {

	// file, err := os.OpenFile("./OurURL.json", os.O_RDONLY, 0777)
	// if err != nil {
	// 	http.Error(w, "Error while opening the file", http.StatusBadRequest)
	// 	return
	// }

	// vers X3
	// 	newkeymap := make(map[string]string)
	// 	fileBytes, _ := os.ReadFile("./OurURL.json")
	// 	err := json.Unmarshal(fileBytes, &newkeymap)
	// 	if err != nil {
	// 		http.Error(w, "Error while opening the file", http.StatusBadRequest)
	// 		return
	// 	}
	// 	originalURL = newkeymap[shortnew]
	// } else {
	// 	var err error
	// 	originalURL, err = repo.GetURL(shortnew)
	// 	if err != nil {
	// 		http.Error(w, "unable to GET Original url", http.StatusBadRequest)
	// 		return
	// 	}
	// }

	//то, что работало:
	// originalURL, err := repo.GetURL(shortnew)
	// if err != nil {
	// 	http.Error(w, "unable to GET Original url", http.StatusBadRequest)
	// 	return
	// }
	// vers X3
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
	// vers 3
	// fileName := configuration.Cfg.FilePath
	// // defer os.Remove(fileName)
	// producer, err := storage.NewProducer(fileName)
	// if err != nil {
	// 	http.Error(w, "Can't create file", http.StatusBadRequest)
	// 	// 	return
	// }
	// defer producer.Close()
	// events := storage.Event{ShortJSON: shortURL.String(), LongJSON: longURL}
	// if err := producer.WriteEvent(&events); err != nil {
	// 	http.Error(w, "Can't WriteEvent()", http.StatusBadRequest)
	// 	return
	// }
	// vers 3.1
	// events := []*storage.Event{{ShortJSON: shortURL.String(), LongJSON: longURL}}
	// for _, event := range events {
	// 	if err := producer.WriteEvent(event); err != nil {
	// 		http.Error(w, "Can't WriteEvent()", http.StatusBadRequest)
	// 		return
	// 	}
	// }
	// vers 2
	// configuration.Cfg.FilePath = "OurURL.json"
	// fileName := configuration.Cfg.FilePath
	// defer os.Remove(fileName)
	// saver, err := storage.NewSaver(fileName)
	// if err != nil {
	// 	http.Error(w, "Can't create saver", http.StatusBadRequest)
	// 	return
	// }
	// defer saver.Close()
	// if err := saver.WriteKeymap(&resobj); err != nil {
	// 	http.Error(w, "Can't save info to the file", http.StatusBadRequest)
	// 	return
	// }
	// vers 1
	//storage.FileWriteFunc(configuration.Cfg.FilePath, shortID.String(), longURL)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&resobj)
}
