package main

import (
	"log"
	"net/http"

	//"os"

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
		r.Get("/{id}", handlers.GetURLHandler)
		r.Post("/", handlers.PostURLHandler)
	})
	return r
}

func main() {
	//os.Setenv("SERVER_ADDRESS", handlers.BaseURL)
	// handlers.Cfg.ServerAddress = ":8080"
	// if err := env.Parse(&handlers.Cfg); err != nil {
	// 	log.Fatal(err)
	// }
	cfg := handlers.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", handlers.GetURLHandler)
		r.Post("/", handlers.PostURLHandler)
		r.Post("/api/shorten", handlers.PostJSONHandler)
	})
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
