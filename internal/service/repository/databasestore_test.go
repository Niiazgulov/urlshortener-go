// Пакет repository, описание в файле doc.go
package repository

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/Niiazgulov/urlshortener.git/internal/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	urlsAmount = 1000
	urlLenght  = 10
)

// Бенчмарк для определения эффективности работы методов AddURL и GetOriginalURL.
func BenchmarkGetURLfunc(b *testing.B) {
	repo := prepareURLFunc()
	urlsToAdd := randURLslice()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	b.ResetTimer()

	for i := 0; i < urlsAmount; i++ {
		_ = repo.AddURL(urlsToAdd[i])
	}

	for i := 0; i < b.N; i++ {
		shortid := urlsToAdd[rand.Intn(urlsAmount)]
		_, err := repo.GetOriginalURL(ctx, shortid.ShortURL)
		require.NoError(b, err)
	}
}

func randomString() string {
	byteURL := make([]byte, urlLenght)
	for i := range byteURL {
		byteURL[i] = Symbols[rand.Intn(urlLenght)]
	}
	return string(byteURL)
}

func randURLslice() []URL {
	urlsToAdd := make([]URL, 0)

	for i := 0; i < urlsAmount; i++ {
		shortID := randomString()
		longURL := randomString()
		userID := randomString()
		u := URL{
			OriginalURL: longURL,
			ShortURL:    shortID,
			UserID:      userID,
		}
		urlsToAdd = append(urlsToAdd, u)
	}

	return urlsToAdd
}

func prepareURLFunc() AddorGetURL {
	cfg, err := configuration.NewConfig()
	if err != nil {
		fmt.Print("Benchmark: can't make a cfg", err)
	}
	repo, err := GetRepository(cfg)
	if err != nil {
		fmt.Print("Benchmark: can't make a repo", err)
	}

	return repo
}

// Структура для теста DB
type DBRepoTest struct {
	suite.Suite
	repo AddorGetURL
}

// Установка параметров
func (s *DBRepoTest) SetupSuite() {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		// dsn = "postgres://postgres:180612@localhost:5432/urldb?sslmode=disable"
		dsn = "postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable"
	}
	repo, err := NewDataBaseStorage(dsn)
	require.NoError(s.T(), err)
	s.repo = repo
}

// TestAddGetURL
func (s *DBRepoTest) TestAddGetURL() {
	user := GenerateRandomIntString()
	testURL := URL{
		OriginalURL: GenerateRandomString(),
		ShortURL:    GenerateRandomString(),
		UserID:      user,
	}
	err := s.repo.AddURL(testURL)
	require.NoError(s.T(), err)
	response, err := s.repo.GetOriginalURL(context.Background(), testURL.ShortURL)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), testURL.OriginalURL, response)
}

// TestGetShortURL
func (s *DBRepoTest) TestGetShortURL() {
	user := GenerateRandomIntString()
	testURL := URL{
		OriginalURL: GenerateRandomString(),
		ShortURL:    GenerateRandomString(),
		UserID:      user,
	}
	err := s.repo.AddURL(testURL)
	require.NoError(s.T(), err)
	response, err := s.repo.GetShortURL(context.Background(), testURL.OriginalURL)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), testURL.ShortURL, response)
}

// TestBatchURL
func (s *DBRepoTest) TestBatchURL() {
	testURL1 := URL{
		OriginalURL:   GenerateRandomString(),
		ShortURL:      GenerateRandomString(),
		CorrelationID: "testcorrid",
	}
	testURL2 := URL{
		OriginalURL:   GenerateRandomString(),
		ShortURL:      GenerateRandomString(),
		CorrelationID: "testcorrid2",
	}
	testURL3 := URL{
		OriginalURL:   GenerateRandomString(),
		ShortURL:      GenerateRandomString(),
		CorrelationID: "testcorrid3",
	}
	_, err := s.repo.BatchURL(context.Background(), []URL{testURL1, testURL2, testURL3})
	require.NoError(s.T(), err)
	response, err := s.repo.GetOriginalURL(context.Background(), testURL1.ShortURL)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), testURL1.OriginalURL, response)

	require.NoError(s.T(), err)
	response, err = s.repo.GetOriginalURL(context.Background(), testURL2.ShortURL)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), testURL2.OriginalURL, response)

	require.NoError(s.T(), err)
	response, err = s.repo.GetOriginalURL(context.Background(), testURL3.ShortURL)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), testURL3.OriginalURL, response)
}

// TestFindAllUserUrls
func (s *DBRepoTest) TestFindAllUserUrls() {
	user := GenerateRandomIntString()
	testURL1 := URL{
		OriginalURL:   GenerateRandomString(),
		ShortURL:      GenerateRandomString(),
		CorrelationID: "testcorrid",
		UserID:        user,
	}
	testURL2 := URL{
		OriginalURL:   GenerateRandomString(),
		ShortURL:      GenerateRandomString(),
		CorrelationID: "testcorrid2",
		UserID:        user,
	}
	testURL3 := URL{
		OriginalURL:   GenerateRandomString(),
		ShortURL:      GenerateRandomString(),
		CorrelationID: "testcorrid3",
		UserID:        user,
	}
	_, err := s.repo.BatchURL(context.Background(), []URL{testURL1, testURL2, testURL3})
	require.NoError(s.T(), err)
	response, err := s.repo.FindAllUserUrls(context.Background(), user)
	assert.NoError(s.T(), err)
	m := make(map[string]string)
	m[testURL1.ShortURL] = testURL1.OriginalURL
	m[testURL2.ShortURL] = testURL2.OriginalURL
	m[testURL3.ShortURL] = testURL3.OriginalURL
	assert.Equal(s.T(), m, response)
}

// TestDeleteUrls
func (s *DBRepoTest) TestDeleteUrls() {
	user := GenerateRandomIntString()
	testURL1 := URL{
		OriginalURL: GenerateRandomString(),
		ShortURL:    GenerateRandomString(),
		UserID:      user,
	}
	testURL2 := URL{
		OriginalURL: GenerateRandomString(),
		ShortURL:    GenerateRandomString(),
		UserID:      user,
	}
	testURL3 := URL{
		OriginalURL: GenerateRandomString(),
		ShortURL:    GenerateRandomString(),
		UserID:      user,
	}
	_, err := s.repo.BatchURL(context.Background(), []URL{testURL1, testURL2, testURL3})
	require.NoError(s.T(), err)
	err = s.repo.DeleteUrls([]URL{testURL1, testURL2, testURL3})
	assert.NoError(s.T(), err)

	_, err = s.repo.GetOriginalURL(context.Background(), testURL1.ShortURL)
	assert.Error(s.T(), ErrURLdeleted, err)
	_, err = s.repo.GetOriginalURL(context.Background(), testURL2.ShortURL)
	assert.Error(s.T(), ErrURLdeleted, err)
	_, err = s.repo.GetOriginalURL(context.Background(), testURL3.ShortURL)
	assert.Error(s.T(), ErrURLdeleted, err)
}

// Запуск теста
func TestDBRepoTest(t *testing.T) {
	suite.Run(t, new(DBRepoTest))
}
