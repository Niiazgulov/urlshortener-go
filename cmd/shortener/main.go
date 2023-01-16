package main

import (
	"log"
	"net/http"

	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/handlers"
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
	log.Fatal(http.ListenAndServe(":8080", r))
}
