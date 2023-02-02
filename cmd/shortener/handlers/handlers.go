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
		for _, err := repo.GetURL(shortID); !errors.Is(err, repository.ErrKeyNotFound); _, err = repo.GetURL(shortID) {
			if err != nil {
				log.Printf("unable to get URL by short ID: %v", err)
				http.Error(w, "unable to get url from DB", http.StatusNetworkAuthenticationRequired) //511
				return
			}
			shortID = repository.GenerateRandomString()
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
		userID, tokenCookie, err := getUserIDfromCookie(repo, r)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError) // 501
			return
		}
		if tokenCookie != nil {
			http.SetCookie(w, tokenCookie)
		}
		ourPoorURL := repository.URL{ShortURL: shortID, OriginalURL: longURL}
		err = repo.AddURL(ourPoorURL, userID)
		if err != nil {
			http.Error(w, "Status internal server error", http.StatusBadRequest)
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
			http.Error(w, "unable to QueryUnescape longURL", http.StatusBadRequest)
			return
		}
		shortID := repository.GenerateRandomString()
		for _, err := repo.GetURL(shortID); !errors.Is(err, repository.ErrKeyNotFound); _, err = repo.GetURL(shortID) {
			if err != nil {
				log.Printf("unable to get URL by short ID: %v", err)
				http.Error(w, "unable to get url from DB", http.StatusInternalServerError) //502
				return
			}
			shortID = repository.GenerateRandomString()
		}
		userID, tokenCookie, err := getUserIDfromCookie(repo, r)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError) //503
			return
		}
		if tokenCookie != nil {
			http.SetCookie(w, tokenCookie)
		}
		ourPoorURL := repository.URL{ShortURL: shortID, OriginalURL: longURL}
		err = repo.AddURL(ourPoorURL, userID)
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

func GetHandler(repo repository.AddorGetURL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortnew := chi.URLParam(r, "id")
		originalURL, err := repo.GetURL(shortnew)
		if err != nil && !errors.Is(err, repository.ErrKeyNotExists) {
			log.Printf("unable to get key from repo: %v", err)
			http.Error(w, "unable to GET Original url", http.StatusInternalServerError) //504
			return
		}
		if errors.Is(err, repository.ErrKeyNotExists) {
			http.Error(w, "unable to GET Original url", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func GetPingHandler(Cfg configuration.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db, err := sql.Open("pgx", configuration.Cfg.DBPath)
		if err != nil {
			http.Error(w, "unable to open DataBase (GetPingHandler)", http.StatusBadRequest)
			return
		}
		defer db.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		err = db.PingContext(ctx)
		if err != nil {
			http.Error(w, "unable to open DataBase (GetPingHandler)", http.StatusBadRequest)
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
}

type UserURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func GetUserAllUrlsHandler(repo repository.AddorGetURL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserID(r)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		urlsmap, err := repo.FindAllUserUrls(userID)
		if err != nil {
			if err == repository.ErrKeyNotFound {
				w.WriteHeader(http.StatusNoContent)
			} else {
				log.Println("Error while getting URLs", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		} else {
			var urlsList []UserURLs
			for urlID, originalURL := range urlsmap {
				urlsList = append(urlsList, UserURLs{ShortURL: configuration.Cfg.BaseURLAddress + "/" + urlID, OriginalURL: originalURL})
			}
			resbyte, err := json.Marshal(urlsList)
			if err != nil {
				log.Println("Error while serializing response", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(resbyte)
		}
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

func getUserIDfromCookie(repo repository.AddorGetURL, r *http.Request) (string, *http.Cookie, error) {
	var cookie *http.Cookie
	sign, err := r.Cookie(userIDCookie)
	if err != nil {
		userID, err := repo.AddNewUser()
		if err != nil {
			log.Println("Error while adding user", err)
			return "", nil, err
		}
		signValue, err := NewUserSign(userID)
		if err != nil {
			log.Println("Error of creating user sign", err)
			return "", nil, err
		}
		cookie := &http.Cookie{Name: userIDCookie, Value: signValue, MaxAge: 0}
		return userID, cookie, nil // added
	} else {
		signValue := sign.Value
		userID, _, err := GetUserSign(signValue)
		if err != nil {
			log.Println("Error when getting of user sign", err)
			return "", nil, err
		}
		return userID, cookie, nil
	}
}

// func getUserIDfromCookie(repo repository.AddorGetURL, r *http.Request) (string, *http.Cookie, error) {
// 	var (
// 		userID    string
// 		checkAuth bool
// 		signValue string
// 		cookie    *http.Cookie
// 	)
// 	sign, err := r.Cookie(userIDCookie)
// 	if err == nil {
// 		signValue = sign.Value
// 		userID, checkAuth, err = GetUserSign(signValue)
// 		if err != nil {
// 			log.Println("Error while checking of sign", err)
// 			return "", nil, err
// 		}
// 	}
// 	if err != nil || !checkAuth {
// 		userID, err = repo.AddNewUser()
// 		if err != nil {
// 			log.Println("Error while adding user", err)
// 			return "", nil, err
// 		}
// 		signValue, err = NewUserSign(userID)
// 		if err != nil {
// 			log.Println("Error while creating of sign", err)
// 			return "", nil, err
// 		}
// 		cookie = &http.Cookie{Name: userIDCookie, Value: signValue, MaxAge: 0}
// 	}
// 	return userID, cookie, nil
// }

func getUserID(r *http.Request) (string, error) {
	encodedCookie, err := r.Cookie(userIDCookie)
	if err != nil {
		return "", err
	}
	userID, checkAuth, err := GetUserSign(encodedCookie.Value)
	if err != nil {
		log.Println("Error while checking of sign", err)
		return "", err
	}
	if !checkAuth {
		return "", repository.ErrIDNotValid
	}
	return userID, nil
}
