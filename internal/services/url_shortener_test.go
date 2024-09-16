package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strconv"
	"testing"
	"time"
	"url-shortener/internal/domain"
	urlMocks "url-shortener/internal/mocks"
	"url-shortener/internal/services/encoder/base62"
	"url-shortener/internal/services/uniqueIdGenerator/go-snowflake-master"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestURLShortener_Create(t *testing.T) {
	t.Run("Create new URL", func(t *testing.T) {
		logger := &slog.Logger{}
		cache := urlMocks.NewCache(t)
		db := urlMocks.NewDatabase(t)
		shortener := New(logger, cache, db)

		destURL := "https://example.com"
		id := snowflake.ID() + 1

		encodedURL := base62.Base62Encode(id)
		expectedURL := &domain.URL{
			Id:       strconv.Itoa(int(id)),
			ShortURL: encodedURL,
			LongURL:  destURL,
		}

		db.On("GetByLongUrl", mock.Anything, destURL).Return(nil, domain.ErrOriginalURLNotFound)
		db.On("InsertUrl", mock.Anything, *expectedURL).Return(nil)
		db.On("GetCountShortUrls", mock.Anything).Return(10, nil)

		actualURL, count, err := shortener.Create(context.Background(), destURL)

		assert.NoError(t, err)
		assert.Equal(t, expectedURL, actualURL)
		assert.Equal(t, 10, count)
		db.AssertExpectations(t)
	})

	t.Run("Existing URL", func(t *testing.T) {
		logger := &slog.Logger{}
		cache := urlMocks.NewCache(t)
		db := urlMocks.NewDatabase(t)
		shortener := New(logger, cache, db)

		destURL := "https://example.com"
		existingURL := &domain.URL{
			Id:       "123",
			ShortURL: "shortURL",
			LongURL:  destURL,
		}

		db.On("GetByLongUrl", mock.Anything, destURL).Return(existingURL, nil)

		actualURL, count, err := shortener.Create(context.Background(), destURL)

		assert.NoError(t, err)
		assert.Equal(t, existingURL, actualURL)
		assert.Equal(t, 0, count)
		db.AssertExpectations(t)
	})

	t.Run("Error on GetByLongUrl", func(t *testing.T) {
		logger := &slog.Logger{}
		cache := urlMocks.NewCache(t)
		db := urlMocks.NewDatabase(t)
		shortener := New(logger, cache, db)

		destURL := "https://example.com"

		db.On("GetByLongUrl", mock.Anything, destURL).Return(nil, errors.New("database error"))

		_, _, err := shortener.Create(context.Background(), destURL)

		assert.Error(t, err)
		db.AssertExpectations(t)
	})

	t.Run("Error on InsertUrl", func(t *testing.T) {
		logger := &slog.Logger{}
		cache := urlMocks.NewCache(t)
		db := urlMocks.NewDatabase(t)
		shortener := New(logger, cache, db)

		destURL := "https://example.com"
		id := snowflake.ID() + 1

		encodedURL := base62.Base62Encode(id)
		expectedURL := &domain.URL{
			Id:       strconv.Itoa(int(id)),
			ShortURL: encodedURL,
			LongURL:  destURL,
		}

		db.On("GetByLongUrl", mock.Anything, destURL).Return(nil, domain.ErrOriginalURLNotFound)
		db.On("InsertUrl", mock.Anything, *expectedURL).Return(errors.New("database error"))

		_, _, err := shortener.Create(context.Background(), destURL)

		assert.Error(t, err)
		db.AssertExpectations(t)
	})

	t.Run("Error on GetCountShortUrls", func(t *testing.T) {
		logger := &slog.Logger{}
		cache := urlMocks.NewCache(t)
		db := urlMocks.NewDatabase(t)
		shortener := New(logger, cache, db)

		destURL := "https://example.com"
		id := snowflake.ID() + 1

		encodedURL := base62.Base62Encode(id)
		expectedURL := &domain.URL{
			Id:       strconv.Itoa(int(id)),
			ShortURL: encodedURL,
			LongURL:  destURL,
		}

		db.On("GetByLongUrl", mock.Anything, destURL).Return(nil, domain.ErrOriginalURLNotFound)
		db.On("InsertUrl", mock.Anything, *expectedURL).Return(nil)
		db.On("GetCountShortUrls", mock.Anything).Return(0, errors.New("database error"))

		_, _, err := shortener.Create(context.Background(), destURL)

		assert.Error(t, err)
		db.AssertExpectations(t)
	})
}

func TestURLShortener_GetOriginalURL(t *testing.T) {
	t.Run("Get from cache", func(t *testing.T) {
		logger := &slog.Logger{}
		cache := urlMocks.NewCache(t)
		db := urlMocks.NewDatabase(t)
		shortener := New(logger, cache, db)

		shortURL := "shortURL"
		longURL := "https://example.com"

		cache.On("Get", mock.Anything, shortURL).Return(longURL, nil)

		actualURL, err := shortener.GetOriginalURL(context.Background(), shortURL)

		assert.NoError(t, err)
		assert.Equal(t, longURL, actualURL)
		cache.AssertExpectations(t)
		db.AssertExpectations(t) // Ensure database is not called
	})

	t.Run("Get from database and cache", func(t *testing.T) {
		logger := &slog.Logger{}
		cache := urlMocks.NewCache(t)
		db := urlMocks.NewDatabase(t)
		shortener := New(logger, cache, db)

		shortURL := "shortURL"
		longURL := "https://example.com"
		expectedURL := &domain.URL{
			LongURL: longURL,
		}

		cache.On("Get", mock.Anything, shortURL).Return(nil, errors.New("cache miss"))
		db.On("GetShortUrl", mock.Anything, shortURL).Return(expectedURL, nil)
		cache.On("Set", mock.Anything, shortURL, expectedURL, time.Hour).Return(nil)

		actualURL, err := shortener.GetOriginalURL(context.Background(), shortURL)

		assert.NoError(t, err)
		assert.Equal(t, longURL, actualURL)
		cache.AssertExpectations(t)
		db.AssertExpectations(t)
	})

	t.Run("Error on database get", func(t *testing.T) {
		logger := &slog.Logger{}
		cache := urlMocks.NewCache(t)
		db := urlMocks.NewDatabase(t)
		shortener := New(logger, cache, db)

		shortURL := "shortURL"

		cache.On("Get", mock.Anything, shortURL).Return(nil, errors.New("cache miss"))
		db.On("GetShortUrl", mock.Anything, shortURL).Return(nil, errors.New("database error"))

		_, err := shortener.GetOriginalURL(context.Background(), shortURL)

		assert.Error(t, err)
		cache.AssertExpectations(t)
		db.AssertExpectations(t)
	})

	t.Run("Error on cache set", func(t *testing.T) {
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
		handler := slog.NewJSONHandler(os.Stdout, opts)
		logger := slog.New(handler)
		cache := urlMocks.NewCache(t)
		db := urlMocks.NewDatabase(t)
		shortener := New(logger, cache, db)

		shortURL := "shortURL"
		longURL := "https://example.com"
		expectedURL := &domain.URL{
			LongURL: longURL,
		}

		cache.On("Get", mock.Anything, shortURL).Return(nil, errors.New("cache miss"))
		db.On("GetShortUrl", mock.Anything, shortURL).Return(expectedURL, nil)
		cache.On("Set", mock.Anything, shortURL, expectedURL, time.Hour).Return(errors.New("cache set error"))

		actualURL, err := shortener.GetOriginalURL(context.Background(), shortURL)

		assert.NoError(t, err)
		assert.Equal(t, longURL, actualURL)
		cache.AssertExpectations(t)
		db.AssertExpectations(t)
	})
}
