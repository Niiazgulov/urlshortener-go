package main

import (
	"log"
	"net/http"
	"os"

	"github.com/NYTimes/gziphandler"
	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/configuration"
	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/handlers"
	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/service/repository"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	if err := env.Parse(&configuration.Cfg); err != nil {
		log.Fatal(err)
	}
	cfg, err := configuration.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	configuration.Cfg = *cfg
	fileTemp, err := os.OpenFile(configuration.Cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatal(err)
	}
	repo, err := repository.NewFileStorage(fileTemp)
	if err != nil {
		log.Fatal(err)
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(gziphandler.GzipHandler)
	r.Use(handlers.DecomprMiddlw)
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", handlers.GetHandler(repo))
		r.Post("/", handlers.PostHandler(repo, configuration.Cfg))
		r.Post("/api/shorten", handlers.PostJSONHandler(repo, configuration.Cfg))
		r.Get("/api/user/urls", handlers.GetUserAllUrlsHandler(repo))
	})
	log.Fatal(http.ListenAndServe(configuration.Cfg.ServerAddress, r))
}
