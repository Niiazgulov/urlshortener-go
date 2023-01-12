package repository

import (
	"errors"
)

var Keymap = map[string]string{}
var ErrorNone = errors.New("short url not found in database")

type URL struct {
	ShortURL    string
	OriginalURL string
}

type AddURLer interface {
	AddURL(shortURL URL) error
}

func (u *URL) AddURL(shortURL URL) error {
	Keymap[u.ShortURL] = u.OriginalURL
	return nil
}

func MakeAdd(a AddURLer, u URL) {
	a.AddURL(u)
}

type GetURLer interface {
	GetURL(shortURL URL) (string, error)
}

func (u *URL) GetURL(shortURL URL) (string, error) {
	ShortNew := Keymap[u.OriginalURL]
	if _, ok := Keymap[u.OriginalURL]; ok {
		return "errURL", ErrorNone
	}
	OriginalURL := Keymap[ShortNew]
	return OriginalURL, nil
}

func MakeGet(g GetURLer, u URL) {
	g.GetURL(u)
}
