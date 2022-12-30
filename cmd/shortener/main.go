package main

import (
	"net/http"
	"urlshortener-1/handlers"
)

func main() {
	// маршрутизация запросов обработчику
	http.HandleFunc("/", handlers.PostHandler)
	http.HandleFunc("/{id}", handlers.GetHandler)
	// // запуск сервера с адресом localhost, порт 8080
	http.ListenAndServe(":8080", nil)
	// log.Fatal(server.ListenAndServe())
}
