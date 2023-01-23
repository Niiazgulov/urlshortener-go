package main

import (
	"log"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/configuration"
	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/handlers"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", handlers.GetHandler)
		r.Post("/", handlers.PostHandler)
	})
	return r
}

func main() {
	if err := env.Parse(&configuration.Cfg); err != nil {
		log.Fatal(err)
	}
	cfg, err := configuration.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	configuration.Cfg = *cfg
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(gziphandler.GzipHandler)
	r.Use(handlers.UncomprMiddlw)
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", handlers.GetHandler)
		r.Post("/", handlers.PostHandler)
		r.Post("/api/shorten", handlers.PostJSONHandler)
	})
	log.Fatal(http.ListenAndServe(configuration.Cfg.ServerAddress, r))
}
