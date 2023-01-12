package handlers

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	. "github.com/Niiazgulov/urlshortener.git/cmd/shortener/service/repository"
	"github.com/go-chi/chi/v5"
)

func GetURLHandler(w http.ResponseWriter, r *http.Request) {
	shortnew := chi.URLParam(r, "id")
	originalURL := Keymap[shortnew]
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func PostURLHandler(w http.ResponseWriter, r *http.Request) {
	short := Encoder()
	if _, ok := Keymap[short]; ok {
		short = Encoder()
	}
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
	Keymap[short] = longURL
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shorturl))
}
