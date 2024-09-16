package services

import (
	"context"
	"url-shortener/internal/domain"
)

type Database interface {
	InsertUrl(ctx context.Context, url domain.URL) error
	GetShortUrl(ctx context.Context, url string) (*domain.URL, error) 
	GetByLongUrl(ctx context.Context, url string) (*domain.URL, error)
	GetCountShortUrls(ctx context.Context) (int, error)
	DeleteShortUrl(ctx context.Context, shortURL string) error
}

type EncoderService interface {
	Base62Encode(number uint64) string
}
