package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
)

type DataBaseStorage struct {
	DataBase *sql.DB
}

func NewDataBaseStorage(databasePath string) (AddorGetURL, error) {
	db, err := sql.Open("pgx", databasePath)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			original_url VARCHAR UNIQUE, 
			short_id VARCHAR UNIQUE,
			id SERIAL PRIMARY KEY,
			user_id VARCHAR)
		`)
	if err != nil {
		return nil, fmt.Errorf("unable to CREATE TABLE in DB: %w", err)
	}
	_, err = db.Exec(`ALTER TABLE urls ADD COLUMN IF NOT EXISTS deleted BOOLEAN NOT NULL DEFAULT false`)
	if err != nil {
		return nil, fmt.Errorf("unable to ADD COLUMN deleted in DB: %w", err)
	}
	return &DataBaseStorage{DataBase: db}, nil
}

func (d *DataBaseStorage) AddURL(u URL) error {
	query := `INSERT INTO urls (original_url, short_id, user_id, deleted) VALUES ($1, $2, $3, $4)`
	_, err := d.DataBase.Exec(query, u.OriginalURL, u.ShortURL, u.UserID, false)
	pgErr := err.(*pq.Error)
	if pgErr.Code == pgerrcode.UniqueViolation {
		return ErrURLexists
	}
	return nil
}

func (d *DataBaseStorage) GetOriginalURL(ctx context.Context, shortid string) (string, error) {
	query := `SELECT original_url, deleted FROM urls WHERE short_id = $1`
	row := d.DataBase.QueryRowContext(ctx, query, shortid)
	var originalURL string
	urlIsDeleted := false
	if err := row.Scan(&originalURL, &urlIsDeleted); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrKeyNotFound
		}
		return "", fmt.Errorf("OMG, I unable to Scan originalURL from DB (GetOriginalURL): %w", err)
	}
	if urlIsDeleted {
		return "", ErrURLdeleted
	}
	return originalURL, nil
}

func (d *DataBaseStorage) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	query := `SELECT short_id FROM urls WHERE original_url = $1`
	row := d.DataBase.QueryRowContext(ctx, query, originalURL)
	var shortURL string
	if err := row.Scan(&shortURL); err != nil {
		if err == sql.ErrNoRows {
			return "", ErrKeyNotFound
		}
		return "", fmt.Errorf("OMG, I unable to Scan shortURL from DB (GetShortURL): %w", err)
	}
	return shortURL, nil
}

func (d *DataBaseStorage) FindAllUserUrls(ctx context.Context, userID string) (map[string]string, error) {
	query := `SELECT original_url, short_id FROM urls WHERE user_id = $1`
	rows, err := d.DataBase.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to return urls from DB (FindAllUserUrls): %w", err)
	}
	defer rows.Close()
	AllIDUrls := make(map[string]string)
	for rows.Next() {
		var shortid string
		var originalURL string
		err = rows.Scan(&originalURL, &shortid)
		if err != nil {
			return nil, err
		}
		AllIDUrls[shortid] = originalURL
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return AllIDUrls, nil
}

func (d *DataBaseStorage) BatchURL(ctx context.Context, userID string, urls []URL) ([]ShortCorrelation, error) {
	var newurls []ShortCorrelation
	for _, batch := range urls {
		shortID := GenerateRandomString()
		shorturl := BaseTest + shortID
		newurl := ShortCorrelation{
			ShortURL:      shorturl,
			CorrelationID: batch.CorrelationID,
		}
		newurls = append(newurls, newurl)
		query := `INSERT INTO urls (original_url, short_id, user_id, deleted) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`
		_, err := d.DataBase.Exec(query, batch.OriginalURL, shortID, userID, false)
		if err != nil {
			var pgerr *pgx.PgError
			if errors.As(err, &pgerr) {
				if pgerr.Code == pgerrcode.UniqueViolation {
					return nil, ErrURLexists
				}
			} else {
				return nil, fmt.Errorf("BatchURL: unable to add URL to DB: %w", err)
			}
		}
	}
	return newurls, nil
}

func (d *DataBaseStorage) DeleteUrls(urls []URL) error {
	if len(urls) == 0 {
		return nil
	}
	deleted := true
	urlsToDelete := make(map[string][]string)
	for _, url := range urls {
		urlsToDelete[url.UserID] = append(urlsToDelete[url.UserID], url.ShortURL)
	}
	query := `UPDATE urls SET deleted =$1 WHERE user_id = $2 AND short_id = any($3)`
	for userID, urlIDs := range urlsToDelete {
		if _, err := d.DataBase.Exec(query, deleted, userID, urlIDs); err != nil {
			return err
		}
	}
	return nil
}

func (d DataBaseStorage) Close() {
	d.DataBase.Close()
}
