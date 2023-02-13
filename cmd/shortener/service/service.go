// Давай сделаем слой с бизнес-логикой, какнибудь назовем его аля service,
// скроем его так же за интерфейсом, сделаем один метод AddURL
// и он будет у нас делать логику с обработкой дублей и генерацией урла,
// находясь при этом между хэндлерами и хранилкой.
// Для получения урла из хранилки будем по-прежнему ходить из хэндлеров в GetOriginalURL напрямую.
// Логика работы с куками, запросом и ответом останется в хэндлерах.

package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Niiazgulov/urlshortener.git/cmd/shortener/service/repository"
)

type ServiceInterf interface {
	AddURL(u repository.URL, userID, shortID string) (string, int, error)
}

type ServiceStruct struct {
	Repos repository.AddorGetURL
}

func (ss *ServiceStruct) AddURL(u repository.URL, userID, shortID string) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	handlerstatus := http.StatusCreated
	err := ss.Repos.AddURL(u, userID)
	if err != nil && !errors.Is(err, repository.ErrURLexists) {
		return "", 0, fmt.Errorf("unable to make repo (Service AddURL): %w", err)
	}
	if errors.Is(err, repository.ErrURLexists) {
		shortID, err = ss.Repos.GetShortURL(ctx, u.OriginalURL)
		if err != nil {
			log.Printf("PostHandler: unable to get shortURL by longURL: %v", err)
			return "", 0, fmt.Errorf("unable to get shortURL from DB (Service AddURL): %w", err)
		}
		handlerstatus = http.StatusConflict
	}
	return shortID, handlerstatus, nil
}
