package repository

import (
	"math/rand"
	"time"
)

var Keymap = map[string]string{}

const (
	symbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	BaseURL = "http://localhost:8080/"
)

func Encoder() string {
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, 0, 7)
	for i := 0; i < 7; i++ {
		s := symbols[rand.Intn(len(symbols))]
		result = append(result, s)
	}
	return string(result)
}
