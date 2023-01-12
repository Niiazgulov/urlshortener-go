package repository

var Keymap = map[string]string{}

type URL struct {
	ShortURL    string
	OriginalURL string
}

type AddURLer interface {
	AddURL(shortURL URL)
}

func (u *URL) AddURL(shortURL URL) {
	Keymap[u.ShortURL] = u.OriginalURL
}

func MakeAdd(a AddURLer, u URL) {
	a.AddURL(u)
}

type GetURLer interface {
	GetURL(shortURL URL) string
}

func (u *URL) GetURL(shortURL URL) string {
	ShortNew := u.ShortURL
	OriginalURL := Keymap[ShortNew]
	return OriginalURL
}

func MakeGet(g GetURLer, u URL) {
	g.GetURL(u)
}
