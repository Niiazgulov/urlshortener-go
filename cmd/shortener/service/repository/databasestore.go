package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx/v5/stdlib"
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
			original_url TEXT UNIQUE, 
			id TEXT,
			user_id TEXT)
		`)
	if err != nil {
		return nil, fmt.Errorf("unable to execute a query to DB: %w", err)
	}
	// _, err = db.Exec(`CREATE UNIQUE INDEX original_unique_idx ON urls (original_url)`)
	// _, err = db.Exec(`ALTER TABLE urls ADD UNIQUE (original_url)`)
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to create unique index to URL in DB: %w", err)
	// }
	return &DataBaseStorage{DataBase: db}, nil
}

func (d *DataBaseStorage) AddURL(u URL, userID string) error {
	query := `INSERT INTO urls (original_url, id, user_id) VALUES ($1, $2, $3)`
	_, err := d.DataBase.Exec(query, u.OriginalURL, u.ShortURL, userID)
	if err != nil && strings.Contains(err.Error(), pgerrcode.UniqueViolation) {
		return ErrURLexists
	}
	// if err != nil {
	// 	var pgerr *pgx.PgError
	// 	if errors.As(err, &pgerr) {
	// 		if pgerr.Code == pgerrcode.UniqueViolation {
	// 			return ErrURLexists
	// 		}
	// 	} else {
	// 		return fmt.Errorf("AddURL: unable to add URL to DB: %w", err)
	// 	}
	// }
	return nil
}

func (d *DataBaseStorage) GetOriginalURL(ctx context.Context, id string) (string, error) {
	query := `SELECT original_url FROM urls WHERE id = $1`
	row := d.DataBase.QueryRowContext(ctx, query, id)
	var originalURL string
	if err := row.Scan(&originalURL); err != nil {
		if err == sql.ErrNoRows {
			return "", ErrKeyNotFound
		}
		return "", fmt.Errorf("OMG, I unable to Scan originalURL from DB (GetOriginalURL): %w", err)
	}
	return originalURL, nil
}

func (d *DataBaseStorage) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	query := `SELECT id FROM urls WHERE original_url = $1`
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

func (d *DataBaseStorage) BatchURL(ctx context.Context, userID string, urls []Correlation) ([]ShortCorrelation, error) {
	var newurls []ShortCorrelation
	for _, batch := range urls {
		shortID := GenerateRandomString()
		shorturl := BaseTest + shortID
		newurl := ShortCorrelation{
			ShortURL:      shorturl,
			CorrelationID: batch.CorrelationID,
		}
		newurls = append(newurls, newurl)
		query := `INSERT INTO urls (original_url, id, user_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
		_, err := d.DataBase.Exec(query, batch.OriginalURL, shortID, userID)
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

func (d DataBaseStorage) Close() {
	d.DataBase.Close()
}
