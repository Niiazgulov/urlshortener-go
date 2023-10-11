package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Niiazgulov/urlshortener-go.git/internal/configuration"
	"github.com/Niiazgulov/urlshortener-go.git/internal/handlers"
	pb "github.com/Niiazgulov/urlshortener-go.git/internal/handlers/proto"
	"github.com/Niiazgulov/urlshortener-go.git/internal/service"
	"github.com/Niiazgulov/urlshortener-go.git/internal/service/repository"
)

// HTTP сервер
func HTTPServer(ctx context.Context, repo repository.AddorGetURL, cfg *configuration.Config, r *chi.Mux) error {
	idleConnsClosed := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	srvShutdown := func(srv *http.Server) {
		select {
		case <-sigint:
		case <-ctx.Done():
		}
		if er := srv.Shutdown(context.Background()); er != nil {
			log.Printf("HTTP server Shutdown: %v", er)
		}
		close(idleConnsClosed)
	}
	if cfg.HTTPS {
		certificate, err := genCertificate()
		if err != nil {
			log.Println(err)
			return err
		}
		server := &http.Server{
			Addr:    configuration.Cfg.ServerAddress,
			Handler: r,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{certificate},
			},
		}
		go srvShutdown(server)
		if err = server.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			log.Printf("HTTP server ListenAndServeTLS: %v", err)
			return err
		}
	} else {
		server := &http.Server{Addr: configuration.Cfg.ServerAddress, Handler: r}
		go srvShutdown(server)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server ListenAndServe: %v", err)
			return err
		}
	}
	<-idleConnsClosed
	fmt.Println("Server Shutdown gracefully")
	return nil
}

// GRPC сервер
func GRPCServer(ctx context.Context, repo repository.AddorGetURL, cfg *configuration.Config, serv service.ServiceStruct) error {
	listen, err := net.Listen("tcp", cfg.ServerGRPC)
	if err != nil {
		log.Println(err)
		return err
	}

	var serverOpts []grpc.ServerOption
	if cfg.GRPCTLS {
		cert, err := genCertificate()
		if err != nil {
			log.Println(err)
			return err
		}
		conf := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		tlsCredentials := credentials.NewTLS(conf)
		serverOpts = []grpc.ServerOption{grpc.Creds(tlsCredentials)}
	}

	s := grpc.NewServer(serverOpts...)
	sigint := make(chan os.Signal, 1)
	connsClosed := make(chan struct{})
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		select {
		case <-sigint:
		case <-ctx.Done():
		}
		s.GracefulStop()
		close(connsClosed)
	}()

	pb.RegisterURLShortenerServer(s, handlers.NewServer(repo, cfg, serv))
	log.Printf("Started gRPC server on %s\n", cfg.ServerGRPC)
	if err := s.Serve(listen); err != nil {
		log.Println(err)
		return err
	}
	<-connsClosed
	log.Printf("Stopped gRPC server on %s\n", cfg.ServerGRPC)
	return nil
}
