package main

import (
	"log"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/Niiazgulov/urlshortener.git/internal/configuration"
	"github.com/Niiazgulov/urlshortener.git/internal/handlers"
	"github.com/Niiazgulov/urlshortener.git/internal/service"
	"github.com/Niiazgulov/urlshortener.git/internal/service/repository"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
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
	repo, err := repository.GetRepository(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()
	serv := service.ServiceStruct{Repos: repo}
	jobCh := make(chan repository.DeleteURLsJob, 200)
	for i := 0; i < cfg.WorkerCount; i++ {
		go func() {
			for job := range jobCh {
				if err := repo.DeleteUrls(job.RequestURLs); err != nil {
					log.Fatal(err)
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
	r.Use(handlers.DecomprMiddlw)
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", handlers.GetHandler(repo))
		r.Get("/api/user/urls", handlers.GetUserAllUrlsHandler(repo))
		r.Get("/ping", handlers.GetPingHandler(repo, configuration.Cfg))
		r.Post("/", handlers.PostHandler(repo, serv, configuration.Cfg))
		r.Post("/api/shorten", handlers.PostJSONHandler(repo, serv, configuration.Cfg))
		r.Post("/api/shorten/batch", handlers.PostBatchHandler(repo))
		r.Delete("/api/user/urls", handlers.DeleteUrlsHandler(repo, jobCh))
	})
	log.Fatal(http.ListenAndServe(configuration.Cfg.ServerAddress, r))
}
