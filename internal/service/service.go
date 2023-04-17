// Пакет service - слой абстракции, в котором выполняется бизнес-логика с обработкой дублей и генерацией URL, находясь при этом между хэндлерами и хранилищем.
package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Niiazgulov/urlshortener.git/internal/service/repository"
)

// Дополнительный интерфейс для работы с методом AddURL.
type ServiceInterf interface {
	AddURL(u repository.URL, userID, shortID string) (string, int, error) // метод AddURL
}

// Структура, в которую передается объект основного интерфейса хранилища.
type ServiceStruct struct {
	Repos repository.AddorGetURL // объект основного интерфейса хранилища
}

// Надстройка основного метода AddURL - проверят URL на наличие в хранилище.
func (ss *ServiceStruct) AddURL(u repository.URL, shortID string) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	handlerstatus := http.StatusCreated
	err := ss.Repos.AddURL(u) // Вызов основного метода AddURL.
	if err != nil && !errors.Is(err, repository.ErrURLexists) {
		return "", 0, fmt.Errorf("unable to make repo (Service AddURL): %w", err)
	}
	if errors.Is(err, repository.ErrURLexists) { // Если URL существует, из хранилища достается связанный с ним shortID.
		shortID, err = ss.Repos.GetShortURL(ctx, u.OriginalURL)
		if err != nil {
			log.Printf("Service AddURL: unable to get shortURL by longURL: %v", err)
			return "", 0, fmt.Errorf("unable to get shortURL from DB (Service AddURL): %w", err)
		}
		handlerstatus = http.StatusConflict // И устанавливается статус 409.
	}
	return shortID, handlerstatus, nil
}
