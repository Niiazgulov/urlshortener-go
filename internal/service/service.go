// Пакет service - слой абстракции, в котором выполняется бизнес-логика с обработкой дублей и генерацией URL, находясь при этом между хэндлерами и хранилищем.
package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Niiazgulov/urlshortener-go.git/internal/service/repository"
)

// Дополнительный интерфейс для работы с методом AddURL.
type ServiceInterf interface {
	AddURL(u repository.URL, userID, shortID string) (string, int, error) // метод AddURL
	GetUserID(sign string) (int, error)
	GetCreateUserID(ctx context.Context, sign string) (int, string, error)
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

func (ss *ServiceStruct) newUserID(ctx context.Context) (string, string, error) {
	userID := repository.GenerateRandomIntString()
	signValue, err := NewUserSign(userID)
	if err != nil {
		log.Println("Error of creating user sign (UserIDfromCookie)", err)
		return "", "", err
	}
	return userID, signValue, err
}

// Функция возвращает userID из подписи
func (ss *ServiceStruct) GetUserID(sign string) (string, error) {
	userID, checkAuth, err := GetUserSign(sign)
	if err != nil {
		log.Println("Error while checking of sign", err)
		return "", err
	}
	if !checkAuth {
		return "", repository.ErrIDNotValid
	}
	return userID, nil
}

// Функция получает или создает userID из подписи
func (ss *ServiceStruct) GetCreateUserID(ctx context.Context, sign string) (string, string, error) {
	if sign == "" {
		ss.newUserID(ctx)
	}
	userID, checkAuth, err := GetUserSign(sign)
	if err != nil {
		log.Println("Error while checking of sign", err)
		return "", "", err
	}
	if !checkAuth {
		return ss.newUserID(ctx)
	}
	return userID, sign, nil
}
