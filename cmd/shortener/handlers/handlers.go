package handlers

import (
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"

	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/configuration"
	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/service/repository"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func PostHandler(repo repository.AddorGetURL, Cfg configuration.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortID := repository.GenerateRandomString()
		// Здесь мы проверяем что урл который мы сгенерировали отсутствует в базе
		// если же он там есть, мы перегенерируем и так пока не получим уникальный
		for _, err := repo.GetURL(r.Context(), shortID); !errors.Is(err, repository.ErrKeyNotFound); _, err = repo.GetURL(r.Context(), shortID) {
			if err != nil {
				log.Printf("PostHandler: unable to get URL by short ID: %v", err)
				http.Error(w, "PostHandler: unable to get url from DB", http.StatusNetworkAuthenticationRequired) //511
				return
			}
			shortID = repository.GenerateRandomString()
		}
		shorturl := configuration.Cfg.ConfigURL.JoinPath(shortID)
		longURLByte, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "PostHandler: can't read Body", http.StatusBadRequest)
			return
		}
		longURL := strings.TrimPrefix(string(longURLByte), "url=")
		longURL, err = url.QueryUnescape(longURL)
		if err != nil {
			http.Error(w, "PostHandler: unable to unescape query in input url", http.StatusBadRequest)
			return
		}
		userID, token, err := UserIDfromCookie(repo, r)
		if err != nil {
			http.Error(w, "PostHandler: Status internal server error", http.StatusInternalServerError) // 501
			return
		}
		if token != nil {
			http.SetCookie(w, token)
		}
		ourPoorURL := repository.URL{ShortURL: shortID, OriginalURL: longURL}
		err = repo.AddURL(ourPoorURL, userID)
		if err != nil {
			http.Error(w, "PostHandler: Status internal server error", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shorturl.String()))
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
			http.Error(w, "PostJSONHandler: unable to QueryUnescape longURL", http.StatusBadRequest)
			return
		}
		shortID := repository.GenerateRandomString()
		for _, err := repo.GetURL(r.Context(), shortID); !errors.Is(err, repository.ErrKeyNotFound); _, err = repo.GetURL(r.Context(), shortID) {
			if err != nil {
				log.Printf("PostJSONHandler: unable to get URL by short ID: %v", err)
				http.Error(w, "PostJSONHandler: unable to get url from DB", http.StatusInternalServerError) //502
				return
			}
			shortID = repository.GenerateRandomString()
		}
		userID, token, err := UserIDfromCookie(repo, r)
		if err != nil {
			http.Error(w, "PostJSONHandler: Status internal server error", http.StatusInternalServerError) //503
			return
		}
		if token != nil {
			http.SetCookie(w, token)
		}
		ourPoorURL := repository.URL{ShortURL: shortID, OriginalURL: longURL}
		err = repo.AddURL(ourPoorURL, userID)
		if err != nil {
			http.Error(w, "PostJSONHandler: Status internal server error", http.StatusBadRequest)
			return
		}
		shortURL := configuration.Cfg.ConfigURL.JoinPath(shortID)
		resobj := repository.JSONKeymap{ShortJSON: shortURL.String(), LongJSON: longURL}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(&resobj)
	}
}

func GetHandler(repo repository.AddorGetURL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortnew := chi.URLParam(r, "id")
		originalURL, err := repo.GetURL(r.Context(), shortnew)
		if err != nil && !errors.Is(err, repository.ErrKeyNotExists) {
			log.Printf("unable to get key from repo: %v", err)
			http.Error(w, "unable to GET Original url (GetHandler)", http.StatusInternalServerError) //504
			return
		}
		if errors.Is(err, repository.ErrKeyNotExists) {
			http.Error(w, "unable to GET Original url (GetHandler)", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

type UserURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func GetUserAllUrlsHandler(repo repository.AddorGetURL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encodedCookie, err := r.Cookie(userIDCookie)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		userID, checkAuth, err := GetUserSign(encodedCookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if !checkAuth {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		urlsmap, err := repo.FindAllUserUrls(r.Context(), userID)
		if err != nil {
			if err == repository.ErrKeyNotFound {
				log.Println("Error while FindAllUserUrls", err)
				w.WriteHeader(http.StatusNoContent)
			} else {
				log.Println("Error while getting URLs", err)
				http.Error(w, "GetUserAllUrlsHandler: Internal server error", http.StatusInternalServerError)
			}
		} else {
			var urlsList []UserURLs
			for short, originalURL := range urlsmap {
				shorturl := configuration.Cfg.ConfigURL.JoinPath(short)
				urlsList = append(urlsList, UserURLs{ShortURL: shorturl.String(), OriginalURL: originalURL})
			}
			resbyte, err := json.Marshal(urlsList)
			if err != nil {
				log.Println("GetUserAllUrlsHandler: Error while serializing response", err)
				http.Error(w, "Status internal server error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(resbyte)
		}
	}
}

func GetPingHandler(repo repository.AddorGetURL, Cfg configuration.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbstorage := repo.(*repository.DataBaseStorage)
		db := dbstorage.DataBase
		var err error
		if db == nil {
			db, err = sql.Open("pgx", configuration.Cfg.DBPath)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			defer db.Close()
		}
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		err = db.PingContext(ctx)
		if err != nil {
			http.Error(w, "bad connection to DataBase (GetPingHandler)", http.StatusInternalServerError)
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func DecomprMiddlw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError) //507
				return
			}
			r.Body = io.NopCloser(gz)
			defer gz.Close()
		}
		next.ServeHTTP(w, r)
	})
}

const userIDCookie = "useridcookie"

func UserIDfromCookie(repo repository.AddorGetURL, r *http.Request) (string, *http.Cookie, error) {
	var cookie *http.Cookie
	sign, err := r.Cookie(userIDCookie)
	if err != nil {
		userID := repository.GenerateRandomIntString()
		signValue, err := NewUserSign(userID)
		if err != nil {
			log.Println("Error of creating user sign (UserIDfromCookie)", err)
			return "", nil, err
		}
		cookie := &http.Cookie{Name: userIDCookie, Value: signValue, MaxAge: 0}
		return userID, cookie, nil // added
	} else {
		signValue := sign.Value
		userID, checkAuth, err := GetUserSign(signValue)
		if err != nil {
			log.Println("Error when getting of user sign (UserIDfromCookie)", err)
			return "", nil, err
		}
		if !checkAuth {
			log.Println("Error of equal checking (UserIDfromCookie)", err)
			return "", nil, err
		}
		return userID, cookie, nil
	}
}
