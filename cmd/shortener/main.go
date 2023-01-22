package main

import (
	"log"
	"net/http"

	// "flag"

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
	configuration.MakeCfgVars(configuration.Cfg.BaseURLAddress, configuration.Cfg.ServerAddress, configuration.Cfg.FilePath)
	// flag.StringVar(&configuration.Cfg.ServerAddress, "a", "", "server adress")
	// flag.StringVar(&configuration.Cfg.BaseURLAddress, "b", "", "base url adress")
	// flag.StringVar(&configuration.Cfg.FilePath, "f", "", "file path")
	// flag.Parse()
	// flag.StringVar(&configuration.FlagServer, "a", ":8080", "SERVER_ADDRESS")
	// flag.StringVar(&configuration.FlagBase, "b", "http://localhost:8080/", "BASE_URL")
	// flag.StringVar(&configuration.FlagFile, "f", "", "FILE_STORAGE_PATH")
	// flag.Parse()
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", handlers.GetHandler)
		r.Post("/", handlers.PostHandler)
		r.Post("/api/shorten", handlers.PostJSONHandler)
	})
	log.Fatal(http.ListenAndServe(configuration.Cfg.ServerAddress, r))
}
