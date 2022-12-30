package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	keymap = make(map[string]string, 100)
	// newMap = make(map[string]interface{})
	// newMap = make(map[string]string, 100)
	// urlgod OurURL
)

const (
	symbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// type OurURL struct {
// 	urlshort string `json:"short"`
// 	urlorigin  string `json:"origin"`
// }

func Encoder(number uint64) string {
	length := len(symbols)
	var encodedBuilder strings.Builder
	encodedBuilder.Grow(10)
	for ; number > 0; number = number / uint64(length) {
		encodedBuilder.WriteByte(symbols[(number % uint64(length))])
	}
	return encodedBuilder.String()
}

func BestHandlerEver(w http.ResponseWriter, r *http.Request) {
	// этот обработчик принимает только запросы, отправленные методом POST
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Only GET or POST requests are allowed!", http.StatusBadRequest)
		return
	}
	//URLid := r.URL.Query().Get("id")
	longURL := r.URL.Path
	if longURL == "" {
		http.Error(w, "This URL is empty", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPost:
		rand.Seed(time.Now().UnixNano())
		randint := rand.Uint64()
		short := Encoder(randint)
		shorturl := "http://localhost:8080/" + short
		keymap[short] = longURL
		jsonData, err := json.Marshal(keymap)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(jsonData))
		jsonFile, err := os.Create("./OurURL.json")
		if err != nil {
			panic(err)
		}
		defer jsonFile.Close()
		jsonFile.Write(jsonData)
		jsonFile.Close()
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shorturl))
	case http.MethodGet:
		// short := r.FormValue("/")
		//short := "tc6lGK8Nbjt"
		//short := r.Header.Get()
		short := r.URL.Path
		var data []byte
		data, _ = ioutil.ReadFile("OurURL.json")
		var m map[string]string
		err := json.Unmarshal(data, &m)
		if err != nil {
			log.Fatal(err)
		}
		//originalURL := m[short]
		w.Header().Set("Location", short)
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		short2 := r.URL.Path
		originalURL2 := keymap[short2]
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Location", originalURL2)
	}
}

func main() {
	// маршрутизация запросов обработчику
	http.HandleFunc("/", BestHandlerEver)
	// http.HandleFunc("/{keymap[shorturl]}", GetHandler)
	// // запуск сервера с адресом localhost, порт 8080
	http.ListenAndServe(":8080", nil)
	// log.Fatal(server.ListenAndServe())
}
