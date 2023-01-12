package repository

var Keymap = map[string]string{}

type URL struct {
	ShortURL    string
	OriginalURL string
}

type AddURLer interface {
	AddURL(shortURL URL) map[string]string
}

func (u *URL) AddURL(shortURL URL) map[string]string {
	key := u.ShortURL
	val := u.OriginalURL
	Keymap[key] = val
	return Keymap
}

func MakeAdd(a AddURLer, u URL) map[string]string {
	result := a.AddURL(u)
	return result
}

type GetURLer interface {
	GetURL(shortURL URL) string
}

func (u *URL) GetURL(shortURL URL) string {
	ShortNew := u.ShortURL
	OldOriginalURL := Keymap[ShortNew]
	return OldOriginalURL
}

func MakeGet(g GetURLer, u URL) string {
	result := g.GetURL(u)
	return result
}
