package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/NYTimes/gziphandler"
	"github.com/Niiazgulov/urlshortener.git/internal/configuration"
	"github.com/Niiazgulov/urlshortener.git/internal/handlers"
	"github.com/Niiazgulov/urlshortener.git/internal/service"
	"github.com/Niiazgulov/urlshortener.git/internal/service/repository"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func NewRouter() chi.Router {
	if err := env.Parse(&configuration.Cfg); err != nil {
		log.Fatal(err)
	}
	fileTemp, err := os.OpenFile("OurURL.json", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatal(err)
	}
	repo, err := repository.NewFileStorage(fileTemp)
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := configuration.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	configuration.Cfg = *cfg
	serv := service.ServiceStruct{Repos: repo}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(gziphandler.GzipHandler)
	r.Use(handlers.DecomprMiddlw)
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", handlers.GetHandler(repo))
		r.Post("/", handlers.PostHandler(repo, serv, configuration.Cfg))
		r.Post("/api/shorten", handlers.PostJSONHandler(repo, serv, configuration.Cfg))
	})
	return r
}

var testkeymap = map[string]string{}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (int, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	if method == http.MethodPost {
		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		defer resp.Body.Close()
		return resp.StatusCode, string(respBody)
	} else {
		respBody := resp.Header.Get("Location")
		require.NoError(t, err)
		defer resp.Body.Close()
		return resp.StatusCode, string(respBody)
	}
}

func TestRouter(t *testing.T) {
	r := NewRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()
	statusCode, body := testRequest(t, ts, "POST", "/")
	testkeymap[body] = ts.URL
	assert.Equal(t, http.StatusCreated, statusCode)
	statusCode, body = testRequest(t, ts, "GET", "/{id}")
	original := testkeymap[body]
	assert.Equal(t, http.StatusInternalServerError, statusCode)
	assert.Equal(t, original, body)
}
