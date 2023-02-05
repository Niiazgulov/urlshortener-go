package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type DataBaseStorage struct {
	DataBase *sql.DB
}

func NewDataBaseStorqage(databasePath string) (*DataBaseStorage, error) {
	db, err := sql.Open("pgx", databasePath)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
		"CREATE TABLE IF NOT EXISTS urls (
			original_url text unique not null, 
			id text unique not null,
			user_id text not null)"
		`)
	if err != nil {
		return nil, fmt.Errorf("unable to execute a query to DB: %w", err)
	}
	return &DataBaseStorage{DataBase: db}, nil
}

func (d DataBaseStorage) AddURL(ctx context.Context, u URL, userID string) error {
	addURLcommand := "INSERT INTO urls (original_url, id, user_id) VALUES ($1, $2, $3)"
	d.DataBase.QueryRowContext(ctx, addURLcommand, u.OriginalURL, u.ShortURL, userID)
	return nil
}

func (d DataBaseStorage) GetURL(ctx context.Context, id string) (string, error) {
	var originalURL string
	getURLcommand := "SELECT original_url FROM urls WHERE id = $1"
	row := d.DataBase.QueryRowContext(ctx, getURLcommand, id)
	err := row.Scan(&originalURL)
	if err != nil {
		return "", fmt.Errorf("unable to Scan originalURL from DB (GetURL): %w", err)
	}
	return originalURL, nil
}

func (d DataBaseStorage) FindAllUserUrls(ctx context.Context, userID string) (map[string]string, error) {
	selectUrls := "SELECT original_url, id FROM urls WHERE user_id = $1"
	rows, err := d.DataBase.QueryContext(ctx, selectUrls, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to return urls from DB (FindAllUserUrls): %w", err)
	}
	defer rows.Close()
	AllIDUrls := make(map[string]string)
	for rows.Next() {
		var id string
		var originalURL string
		err = rows.Scan(&originalURL, &id)
		if err != nil {
			return nil, err
		}
		AllIDUrls[id] = originalURL
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return AllIDUrls, nil
}

func (d DataBaseStorage) Close() {
	d.DataBase.Close()
}