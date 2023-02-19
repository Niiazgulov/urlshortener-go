package handlers

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Niiazgulov/urlshortener.git/internal/configuration"
	"github.com/Niiazgulov/urlshortener.git/internal/service"
	"github.com/Niiazgulov/urlshortener.git/internal/service/repository"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	handlerstatus int
	newshortID    string
)

func PostHandler(repo repository.AddorGetURL, Cfg configuration.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortID := repository.GenerateRandomString()
		shortID = repository.RandomStringUniqueCheck(repo, w, r, shortID)
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
			http.Error(w, "PostHandler: Status internal server error", http.StatusInternalServerError)
			return
		}
		if token != nil {
			http.SetCookie(w, token)
		}
		ourPoorURL := repository.URL{ShortURL: shortID, OriginalURL: longURL, UserID: userID}
		serv := service.ServiceStruct{Repos: repo}
		newshortID, handlerstatus, err = serv.AddURL(ourPoorURL, shortID)
		if err != nil {
			http.Error(w, "PostHandler: Status internal server error", http.StatusInternalServerError)
			return
		}
		shorturl := configuration.Cfg.ConfigURL.JoinPath(newshortID)
		response := shorturl.String()
		w.WriteHeader(handlerstatus)
		w.Write([]byte(response))
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
		shortID = repository.RandomStringUniqueCheck(repo, w, r, shortID)
		userID, token, err := UserIDfromCookie(repo, r)
		if err != nil {
			http.Error(w, "PostJSONHandler: Status internal server error", http.StatusInternalServerError)
			return
		}
		if token != nil {
			http.SetCookie(w, token)
		}
		ourPoorURL := repository.URL{ShortURL: shortID, OriginalURL: longURL, UserID: userID}
		serv := service.ServiceStruct{Repos: repo}
		newshortID, handlerstatus, err = serv.AddURL(ourPoorURL, shortID)
		if err != nil {
			http.Error(w, "PostHandler: Status internal server error", http.StatusInternalServerError)
			return
		}
		shortURL := configuration.Cfg.ConfigURL.JoinPath(newshortID)
		response := repository.JSONKeymap{
			ShortJSON: shortURL.String(),
			LongJSON:  longURL,
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(handlerstatus)
		json.NewEncoder(w).Encode(&response)
	}
}

func GetHandler(repo repository.AddorGetURL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handlerstatus := http.StatusTemporaryRedirect
		shortnew := chi.URLParam(r, "id")
		originalURL, err := repo.GetOriginalURL(r.Context(), shortnew)
		if err != nil && !errors.Is(err, repository.ErrKeyNotExists) && !errors.Is(err, repository.ErrURLdeleted) {
			log.Printf("GetHandler: unable to Get Original url from repo: %v", err)
			http.Error(w, "GetHandler: unable to GET Original url", http.StatusInternalServerError)
			return
		}
		if errors.Is(err, repository.ErrKeyNotExists) {
			http.Error(w, "GetHandler: url not found", http.StatusBadRequest)
			return
		}
		if errors.Is(err, repository.ErrURLdeleted) {
			handlerstatus = http.StatusGone
		}
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Header().Set("Location", originalURL)
		w.WriteHeader(handlerstatus)
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
		userID, _, err := GetUserSign(encodedCookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		urlsmap, err := repo.FindAllUserUrls(r.Context(), userID)
		if err != nil && !errors.Is(err, repository.ErrKeyNotFound) {
			log.Println("Error while getting URLs", err)
			http.Error(w, "GetUserAllUrlsHandler: Internal server error", http.StatusInternalServerError)
			return
		}
		if errors.Is(err, repository.ErrKeyNotFound) {
			log.Println("Error while FindAllUserUrls", err)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		var urlsList []UserURLs
		for short, originalURL := range urlsmap {
			shorturl := configuration.Cfg.ConfigURL.JoinPath(short)
			urlsList = append(urlsList, UserURLs{ShortURL: shorturl.String(), OriginalURL: originalURL})
		}
		response, err := json.Marshal(urlsList)
		if err != nil {
			log.Println("GetUserAllUrlsHandler: Error while serializing response", err)
			http.Error(w, "Status internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	}
}

func GetPingHandler(repo repository.AddorGetURL, Cfg configuration.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbstorage, ok := repo.(*repository.DataBaseStorage)
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		db := dbstorage.DataBase
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		err := db.PingContext(ctx)
		if err != nil {
			http.Error(w, "bad connection to DataBase (GetPingHandler)", http.StatusInternalServerError)
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func PostBatchHandler(repo repository.AddorGetURL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "PostBatchHandler: can't read r.Body", http.StatusBadRequest)
			return
		}
		userID, token, err := UserIDfromCookie(repo, r)
		if err != nil {
			http.Error(w, "PostBatchHandler: Error when getting of userID", http.StatusInternalServerError)
			return
		}
		var originalurls []repository.URL
		err = json.Unmarshal(request, &originalurls)
		if err != nil {
			http.Error(w, "PostBatchHandler: can't Unmarshal request", http.StatusBadRequest)
			return
		}
		result, err := repo.BatchURL(r.Context(), userID, originalurls)
		if err != nil {
			http.Error(w, "PostBatchHandler: Status internal server error (BatchURL)", http.StatusInternalServerError)
			return
		}
		response, err := json.Marshal(result)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if token != nil {
			http.SetCookie(w, token)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(response)
	}
}

func DeleteUrlsHandler(repo repository.AddorGetURL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "DeleteUrlsHandler: can't read r.Body", http.StatusBadRequest)
			return
		}
		userID, token, err := UserIDfromCookie(repo, r)
		if err != nil {
			http.Error(w, "DeleteUrlsHandler: Error when getting of userID", http.StatusInternalServerError)
			return
		}
		var requestURLs []string
		err = json.Unmarshal(request, &requestURLs)
		if err != nil {
			http.Error(w, "DeleteUrlsHandler: can't Unmarshal request", http.StatusBadRequest)
			return
		}
		// go repo.DeleteUrls(r.Context(), userID, requestURLs)
		go repository.DeleteUrlsFunc(repo, requestURLs, userID)
		if token != nil {
			http.SetCookie(w, token)
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

func DecomprMiddlw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
		return userID, cookie, nil
	}
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
