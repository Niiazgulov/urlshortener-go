// Пакет handlers, описание в файле doc.go
package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/NYTimes/gziphandler"
	"github.com/Niiazgulov/urlshortener-go.git/internal/configuration"
	"github.com/Niiazgulov/urlshortener-go.git/internal/service"
	"github.com/Niiazgulov/urlshortener-go.git/internal/service/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
)

// func NewRouter()
func NewRouter() chi.Router {
	cfg := &configuration.Config{
		BaseURLAddress: "http://localhost:8080",
		ServerAddress:  ":8080",
		DBPath:         "postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable",
		// DBPath: "postgres://postgres:180612@localhost:5432/urldb?sslmode=disable",
	}
	cfg.ConfigURL, _ = url.Parse(cfg.BaseURLAddress)
	repo, err := repository.NewDataBaseStorage(cfg.DBPath)
	if err != nil {
		fmt.Print("TestPostHandler: unable to make repo: ", err)
	}
	serv := service.ServiceStruct{Repos: repo}
	jobCh := make(chan repository.DeleteURLsJob, 200)
	for i := 0; i < configuration.Cfg.WorkerCount; i++ {
		go func() {
			for job := range jobCh {
				if err := repo.DeleteUrls(job.RequestURLs); err != nil {
					log.Println("Error while DeleteUrls", err)
				}
			}
		}()
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(gziphandler.GzipHandler)
	r.Use(DecomprMiddlw)
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", GetHandler(repo))
		r.Get("/api/user/urls", GetUserAllUrlsHandler(repo))
		r.Get("/ping", GetPingHandler(repo, *cfg))
		r.Post("/", PostHandler(repo, serv, *cfg))
		r.Post("/api/shorten", PostJSONHandler(repo, serv, *cfg))
		r.Post("/api/shorten/batch", PostBatchHandler(repo))
		r.Delete("/api/user/urls", DeleteUrlsHandler(repo, jobCh))
	})
	return r
}

// TestPostHandler
func TestPostHandler(t *testing.T) {
	longURL := repository.GenerateRandomString()
	tests := []struct {
		name               string
		inputBody          string
		expectedStatusCode int
	}{
		{
			name:               "OK",
			inputBody:          "url=" + longURL,
			expectedStatusCode: 201,
		},
		{
			name:               "Empty input",
			inputBody:          "",
			expectedStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			ts := httptest.NewServer(r)
			defer ts.Close()
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.inputBody))
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatusCode, w.Code)
		})
	}
}

// TestPostJSONHandler
func TestPostJSONHandler(t *testing.T) {
	longURL := repository.GenerateRandomString()
	tests := []struct {
		name               string
		inputBody          string
		expectedStatusCode int
	}{
		{
			name:               "OK",
			inputBody:          "{\"url\":\"" + longURL + "\"}",
			expectedStatusCode: 201,
		},
		{
			name:               "Empty input",
			inputBody:          "",
			expectedStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			ts := httptest.NewServer(r)
			defer ts.Close()
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(tt.inputBody))
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatusCode, w.Code)
		})
	}
}

// TestGetHandler
func TestGetHandler(t *testing.T) {
	longURL := repository.GenerateRandomString()
	tests := []struct {
		name               string
		inputBody          string
		expectedStatusCode int
		expectedLongURL    string
		shortID            string
	}{
		{
			name:               "OK",
			inputBody:          "url=" + longURL,
			expectedStatusCode: 307,
			expectedLongURL:    longURL,
		},
		{
			name:               "Wrong shortID",
			inputBody:          "url=" + repository.GenerateRandomString(),
			expectedStatusCode: 500,
			expectedLongURL:    "",
			shortID:            "wrongshortid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			ts := httptest.NewServer(r)
			defer ts.Close()
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.inputBody))
			r.ServeHTTP(w, req)
			shortIDByte, _ := io.ReadAll(w.Body)
			shortID := strings.TrimPrefix(string(shortIDByte), "http://localhost:8080/")
			if tt.shortID == "" {
				tt.shortID = shortID
			}
			w2 := httptest.NewRecorder()
			req2 := httptest.NewRequest(http.MethodGet, "/"+tt.shortID, nil)
			q := req2.URL.Query()
			q.Add("id", tt.shortID)
			req2.URL.RawQuery = q.Encode()
			r.ServeHTTP(w2, req2)
			assert.Equal(t, tt.expectedStatusCode, w2.Code)
			assert.Equal(t, tt.expectedLongURL, w2.Header().Get("Location"))
		})
	}
}

// TestPostBatchHandler
func TestPostBatchHandler(t *testing.T) {
	tests := []struct {
		name               string
		inputBody          string
		expectedStatusCode int
		expectedResponse   []repository.ShortCorrelation
	}{
		{
			name:               "OK",
			inputBody:          "[{\"correlation_id\":\"corrid1\",\"original_url\":\"" + repository.GenerateRandomString() + "\"},{\"correlation_id\":\"corrid2\",\"original_url\":\"" + repository.GenerateRandomString() + "\"}]",
			expectedResponse:   []repository.ShortCorrelation{{CorrelationID: "corrid1"}, {CorrelationID: "corrid2"}},
			expectedStatusCode: 201,
		},
		{
			name:               "Empty",
			inputBody:          "",
			expectedResponse:   []repository.ShortCorrelation(nil),
			expectedStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			ts := httptest.NewServer(r)
			defer ts.Close()
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString(tt.inputBody))
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			shortIDByte, _ := io.ReadAll(w.Body)
			var shorturls []repository.ShortCorrelation
			json.Unmarshal(shortIDByte, &shorturls)
			for i := range shorturls {
				tt.expectedResponse[i].ShortURL = shorturls[i].ShortURL
			}
			assert.Equal(t, tt.expectedResponse, shorturls)
		})
	}
}

// TestDeleteUrlsHandler
func TestDeleteUrlsHandler(t *testing.T) {
	tests := []struct {
		name               string
		inputBody          string
		expectedStatusCode int
	}{
		{
			name:               "OK",
			inputBody:          "[{\"correlation_id\":\"corrid1\",\"original_url\":\"" + repository.GenerateRandomString() + "\"},{\"correlation_id\":\"corrid2\",\"original_url\":\"" + repository.GenerateRandomString() + "\"}]",
			expectedStatusCode: 202,
		},
		{
			name:               "Empty",
			inputBody:          "",
			expectedStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			ts := httptest.NewServer(r)
			defer ts.Close()
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString(tt.inputBody))
			r.ServeHTTP(w, req)

			shortIDByte, _ := io.ReadAll(w.Body)
			var shorturls []repository.ShortCorrelation
			json.Unmarshal(shortIDByte, &shorturls)
			var deleteString string
			if shorturls != nil {
				var deleteInput []string
				for i := range shorturls {
					shortID := strings.TrimPrefix(string(shorturls[i].ShortURL), "http://localhost:8080/")
					shortID = "\"" + shortID + "\""
					deleteInput = append(deleteInput, shortID)
				}
				deleteString = strings.Join(deleteInput, ", ")
				deleteString = "[" + deleteString + "]"
			}
			w = httptest.NewRecorder()
			req = httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBufferString(deleteString))
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
		})
	}
}
