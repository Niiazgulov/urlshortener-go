package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type DataBaseStorage struct {
	DataBase *sql.DB
}

func NewDataBaseStorqage(databasePath string) (AddorGetURL, error) {
	db, err := sql.Open("pgx", databasePath)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			original_url VARCHAR, 
			id VARCHAR,
			user_id VARCHAR)
		`)
	if err != nil {
		return nil, fmt.Errorf("unable to execute a query to DB: %w", err)
	}
	return &DataBaseStorage{DataBase: db}, nil
}

func (d *DataBaseStorage) AddURL(u URL, userID string) error {
	query := `INSERT INTO urls (original_url, id, user_id) VALUES ($1, $2, $3)`
	_, err := d.DataBase.Exec(query, u.OriginalURL, u.ShortURL, userID)
	if err != nil {
		return fmt.Errorf("unable to AddURL to DB: %w", err)
	}
	return nil
}

func (d *DataBaseStorage) GetURL(ctx context.Context, id string) (string, error) {
	query := `SELECT original_url FROM urls WHERE id = $1`
	row := d.DataBase.QueryRowContext(ctx, query, id)
	var originalURL string
	if err := row.Scan(&originalURL); err != nil {
		return "", fmt.Errorf("OMG, I unable to Scan originalURL from DB (GetURL): %w", err)
	}
	return originalURL, nil
}

func (d *DataBaseStorage) FindAllUserUrls(ctx context.Context, userID string) (map[string]string, error) {
	query := `SELECT original_url, id FROM urls WHERE user_id = $1`
	rows, err := d.DataBase.QueryContext(ctx, query, userID)
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
