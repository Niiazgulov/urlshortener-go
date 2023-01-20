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

	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/service/repository"
	"github.com/go-chi/chi/v5"
)

var (
	repo repository.AddorGetURL
	Cfg  Config
)

func init() {
	repo = repository.NewMemoryRepository()
}

const (
	symbols        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	BaseURL        = "http://localhost:8080/"
	ShortURLMaxLen = 7
)

type Config struct {
	ServerAddress  string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURLAddress string `env:"BASE_URL" envDefault:"http://localhost:8080/"`
	FilePath       string `env:"FILE_STORAGE_PATH" envDefault:"./OurURL.json"`
}

type JSONKeymap struct {
	ShortJSON string `json:"result,omitempty"`
	LongJSON  string `json:"url,omitempty"`
}

// type response1 struct {
// 	Result string `json:"result,omitempty"`
// }

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
	short := shortParse.String()
	BaseCfgURL, _ := url.Parse(Cfg.BaseURLAddress)
	shorturl := BaseCfgURL.JoinPath(short)
	// shorturl := Cfg.BaseURLAddress + short
	// shorturl := BaseURL + short
	// fmt.Println(shorturl)
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
	file, err := os.OpenFile("./OurURL.json", os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		http.Error(w, "error while opening the file", http.StatusBadRequest)
		return
	}
	urlmap := make(map[string]string)
	urlmap[short] = longURL
	jsonData, err := json.Marshal(urlmap)
	if err != nil {
		http.Error(w, "unable to Marshal urlmap", http.StatusBadRequest)
		return
	}
	file.Write(jsonData)
	defer file.Close()
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shorturl.String()))
}

func GetURLHandler(w http.ResponseWriter, r *http.Request) {
	shortnew := chi.URLParam(r, "id")
	// var originalURL string
	// if Cfg.FilePath != "" {
	// 	// file, err := os.OpenFile("./OurURL.json", os.O_RDONLY, 0777)
	// 	// if err != nil {
	// 	// 	http.Error(w, "Error while opening the file", http.StatusBadRequest)
	// 	// 	return
	// 	// }
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
	//var tempmap map[string]string
	var tempStrorage JSONKeymap
	if err := json.NewDecoder(r.Body).Decode(&tempStrorage); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	longURL, err := url.QueryUnescape(tempStrorage.LongJSON)
	//longURL, err := url.QueryUnescape(tempmap["url"])
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
	BaseCfgURL, _ := url.Parse(Cfg.BaseURLAddress)
	shortURL := BaseCfgURL.JoinPath(shortID.String())
	//fmt.Println(shortURL.String())
	resobj := JSONKeymap{ShortJSON: shortURL.String()}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&resobj)

	// longURLByte, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	http.Error(w, "can't read Body", http.StatusBadRequest)
	// 	return
	// }
	// JSONrequest := JSONKeymap{}
	// if err := json.Unmarshal([]byte(longURLByte), &JSONrequest); err != nil {
	// 	http.Error(w, "unable to unescape JSON query in input url", http.StatusBadRequest)
	// 	return
	// }
	// longURL := JSONrequest.LongJSON
	// short := generateRandomString()
	// for _, err := repo.GetURL(short); err == nil; _, err = repo.GetURL(short) {
	// 	short = generateRandomString()
	// }
	// shorturl := Cfg.BaseURLAddress + short
	// JSONresponse := JSONKeymap{ShortJSON: shorturl}
	// response, err := json.Marshal(JSONresponse)
	// if err != nil {
	// 	http.Error(w, "Can't make a json.Marshal operation", http.StatusBadRequest)
	// 	return
	// }
	// ourPoorURL := repository.URL{ShortURL: short, OriginalURL: longURL}
	// err = repo.AddURL(ourPoorURL)
	// if err != nil {
	// 	http.Error(w, "Status internal server error", http.StatusBadRequest)
	// 	return
	// }
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
	// w.Header().Add("Content-Type", "application/json")
	// w.WriteHeader(http.StatusCreated)
	// w.Write(response)
}

// func ShortenHandler(w http.ResponseWriter, r *http.Request) {
// 	var s map[string]string
// 	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	homeURL, _ := url.Parse(Cfg.BaseURLAddress)
// 	longURL, err := url.QueryUnescape(s["url"])
// 	if err != nil {
// 		http.Error(w, "unable to Marshal urlmap", http.StatusBadRequest)
// 		return
// 	}
// 	shortID, err := url.Parse(generateRandomString())
// 	if err != nil {
// 		http.Error(w, "unable to Marshal urlmap", http.StatusBadRequest)
// 		return
// 	}
// 	m := make(map[string]string)
// 	shortURL := homeURL.JoinPath(shortID.String())
// 	if _, ok := m.counters[shortID.String()]; ok {
// 		return
// 	}
// 	m.counters[shortID.String()] = longURL
// 	w.Header().Set("content-type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	obj := response1{
// 		Result: shortURL.String(),
// 	}
// 	json.NewEncoder(w).Encode(&obj)
// }
