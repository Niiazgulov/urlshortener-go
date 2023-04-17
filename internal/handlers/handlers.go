// Пакет handlers, описание в файле doc.go
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

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/Niiazgulov/urlshortener-go.git/internal/configuration"
	"github.com/Niiazgulov/urlshortener-go.git/internal/service"
	"github.com/Niiazgulov/urlshortener-go.git/internal/service/repository"
)

// PostHandler - обработчик эндпоинта POST "/" - добавление в хранилище оригинального URL.
func PostHandler(repo repository.AddorGetURL, serv service.ServiceStruct, Cfg configuration.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortID := repository.GenerateRandomString()                      // Генерирует рандомный shortID.
		shortID = repository.RandomStringUniqueCheck(repo, w, r, shortID) // Проверка shortID на уникальность.
		longURLByte, err := io.ReadAll(r.Body)                            // Считывает с Body longURLByte.
		if err != nil {
			http.Error(w, "PostHandler: can't read Body", http.StatusBadRequest)
			return
		}
		longURL := strings.TrimPrefix(string(longURLByte), "url=") // Преобразовывает []byte в string, отсекает "url=".
		longURL, err = url.QueryUnescape(longURL)
		if err != nil {
			http.Error(w, "PostHandler: unable to unescape query in input url", http.StatusBadRequest)
			return
		}
		if longURL == "" {
			http.Error(w, "PostHandler: empty input url", http.StatusBadRequest)
			return
		}
		userID, token, err := UserIDfromCookie(repo, r) // Получает userID и token.
		if err != nil {
			http.Error(w, "PostHandler: Status internal server error", http.StatusNotImplemented)
			return
		}
		if token != nil {
			http.SetCookie(w, token)
		}
		ourPoorURL := repository.URL{ShortURL: shortID, OriginalURL: longURL, UserID: userID} // Объединяет все элементы в объект структуры URL.
		newshortID, handlerstatus, err := serv.AddURL(ourPoorURL, shortID)                    // С помощью вспомогательного интерфейса добавляет информацию об URL в хранилище.
		if err != nil {
			http.Error(w, "PostHandler: Status internal server error", http.StatusBadRequest)
			return
		}
		shorturl := Cfg.ConfigURL.JoinPath(newshortID)
		response := shorturl.String() // Формирование ответа
		w.WriteHeader(handlerstatus)
		w.Write([]byte(response))
	}
}

// PostJSONHandler - обработчик эндпоинта POST "/api/shorten" - добавление в хранилище оригинального URL из JSON.
func PostJSONHandler(repo repository.AddorGetURL, serv service.ServiceStruct, Cfg configuration.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tempStrorage repository.JSONKeymap // Создание объекта структуры JSONKeymap для хранения информации из Body.
		if err := json.NewDecoder(r.Body).Decode(&tempStrorage); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		longURL, err := url.QueryUnescape(tempStrorage.LongJSON) // Получение longURL string из объекта структуры JSONKeymap.
		if err != nil {
			http.Error(w, "PostJSONHandler: unable to QueryUnescape longURL", http.StatusBadRequest)
			return
		}
		if longURL == "" {
			http.Error(w, "PostJSONHandler: empty input url", http.StatusBadRequest)
			return
		}
		shortID := repository.GenerateRandomString()                      // Генерирует рандомный shortID.
		shortID = repository.RandomStringUniqueCheck(repo, w, r, shortID) // Проверка shortID на уникальность.
		userID, token, err := UserIDfromCookie(repo, r)                   // Получает userID и token.
		if err != nil {
			http.Error(w, "PostJSONHandler: Status internal server error", http.StatusInternalServerError)
			return
		}
		if token != nil {
			http.SetCookie(w, token)
		}
		ourPoorURL := repository.URL{ShortURL: shortID, OriginalURL: longURL, UserID: userID} // Объединяет все элементы в объект структуры URL.
		newshortID, handlerstatus, err := serv.AddURL(ourPoorURL, shortID)                    // С помощью вспомогательного интерфейса добавляет информацию об URL в хранилище.
		if err != nil {
			http.Error(w, "PostHandler: Status internal server error", http.StatusInternalServerError)
			return
		}
		shortURL := Cfg.ConfigURL.JoinPath(newshortID)
		response := repository.JSONKeymap{ // Формирование ответа
			ShortJSON: shortURL.String(),
			LongJSON:  longURL,
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(handlerstatus)
		json.NewEncoder(w).Encode(&response)
	}
}

// GetHandler - обработчик эндпоинта GET "/{id}" - получение оригинального URL из хранилища по shortID.
func GetHandler(repo repository.AddorGetURL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handlerstatus := http.StatusTemporaryRedirect
		shortnew := chi.URLParam(r, "id")                              // Считывание shortID из http.Request
		originalURL, err := repo.GetOriginalURL(r.Context(), shortnew) // Получение оригинального URL из хранилища
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

// UserURLs - структура лоя хранения shortID и original URL.
type UserURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// GetUserAllUrlsHandler - обработчик эндпоинта GET "/api/user/urls" - получение всех URL одного пользователя.
func GetUserAllUrlsHandler(repo repository.AddorGetURL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encodedCookie, err := r.Cookie(userIDCookie) // Получение куки из запроса.
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		userID, _, err := GetUserSign(encodedCookie.Value) // Получение userID из куки.
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		urlsmap, err := repo.FindAllUserUrls(r.Context(), userID) // Получение всех URL одного userID и добавление их в map.
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
		for short, originalURL := range urlsmap { // Преобразование map[string]string в слайс объектов структуры UserURLs.
			shorturl := configuration.Cfg.ConfigURL.JoinPath(short)
			urlsList = append(urlsList, UserURLs{ShortURL: shorturl.String(), OriginalURL: originalURL})
		}
		response, err := json.Marshal(urlsList) // Формирование ответа в JSON.
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

// GetPingHandler - обработчик эндпоинта GET "/ping" - проверка соединения с базой данных.
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

// PostBatchHandler - обработчик эндпоинта POST "/api/shorten/batch" - работа с batch.
func PostBatchHandler(repo repository.AddorGetURL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, err := io.ReadAll(r.Body) // Считывание информации из тела запроса.
		if err != nil {
			http.Error(w, "PostBatchHandler: can't read r.Body", http.StatusBadRequest)
			return
		}
		// _, token, err := UserIDfromCookie(repo, r) // Получает userID и token.
		userID, token, err := UserIDfromCookie(repo, r) // Получает userID и token.
		if err != nil {
			http.Error(w, "PostBatchHandler: Error when getting of userID", http.StatusInternalServerError)
			return
		}
		var originalurls []repository.URL
		err = json.Unmarshal(request, &originalurls) // Считывание информации из JSON в слайс объектов структуры URL.
		if err != nil {
			http.Error(w, "PostBatchHandler: can't Unmarshal request", http.StatusBadRequest)
			return
		}
		if originalurls == nil {
			http.Error(w, "PostBatchHandler: empty input url", http.StatusBadRequest)
			return
		}
		for i := range originalurls {
			originalurls[i].UserID = userID
		}
		result, err := repo.BatchURL(r.Context(), originalurls) // Пролучение списка shortID по CorrelationID.
		// result, err := repo.BatchURL(r.Context(), userID, originalurls) // Пролучение списка shortID по CorrelationID.
		if err != nil {
			http.Error(w, "PostBatchHandler: Status internal server error (BatchURL)", http.StatusInternalServerError)
			return
		}
		response, err := json.Marshal(result) // Формирование ответа в JSON.
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

// DeleteUrlsHandler - обработчик эндпоинта POST "/api/user/urls" - удаление URL пользователя.
func DeleteUrlsHandler(repo repository.AddorGetURL, jobCh chan repository.DeleteURLsJob) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, err := io.ReadAll(r.Body) // Считывание информации из тела запроса.
		if err != nil {
			http.Error(w, "DeleteUrlsHandler: can't read r.Body", http.StatusBadRequest)
			return
		}
		userID, _, err := UserIDfromCookie(repo, r) // Получает userID.
		if err != nil {
			http.Error(w, "DeleteUrlsHandler: Error when getting of userID", http.StatusInternalServerError)
			return
		}
		var requestURLs []string
		err = json.Unmarshal(request, &requestURLs) // Считывание информации из JSON в слайс строк.
		if err != nil {
			http.Error(w, "DeleteUrlsHandler: can't Unmarshal request", http.StatusBadRequest)
			return
		}
		structURLs := make([]repository.URL, 0, len(requestURLs))
		for _, url := range requestURLs { // Считывание информации в слайс объектов структуры URL.
			v := repository.URL{ShortURL: url, UserID: userID}
			structURLs = append(structURLs, v)
		}
		jobCh <- repository.DeleteURLsJob{RequestURLs: structURLs} // Отправка задания на удаление урлов в канал.
		w.WriteHeader(http.StatusAccepted)
	}
}

// Функция для распаковки данных.
func DecomprMiddlw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
			r.Body = io.NopCloser(gz)
			defer gz.Close()
		}
		next.ServeHTTP(w, r)
	})
}

const userIDCookie = "useridcookie"

// Функция для получение userID и токена из куки.
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
