package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"
	"url-shortener/internal/domain"
	"url-shortener/internal/services/encoder/base62"
	"url-shortener/internal/services/uniqueIdGenerator/go-snowflake-master"
	"url-shortener/pkg/cache"
)

type URLShortener struct {
	logger *slog.Logger
	cache  cache.Cache
	db     Database
}

func New(logger *slog.Logger, cache cache.Cache, db Database) *URLShortener {
	return &URLShortener{
		logger: logger,
		cache:  cache,
		db:     db,
	}
}

func (u *URLShortener) Create(ctx context.Context, destUrl string) (*domain.URL, int,  error) {

	// check if link already exists on database
	existUrl, err := u.db.GetByLongUrl(ctx, destUrl)
	if err == nil {
		return existUrl, 0, nil
	}

	if !errors.Is(err, domain.ErrOriginalURLNotFound) {
		return nil, 0, err
	}

	id := snowflake.ID()

	encodedUrl := base62.Base62Encode(id)

	url := domain.URL{
		Id:       strconv.Itoa(int(id)),
		ShortURL: encodedUrl,
		LongURL:  destUrl,
	}

	// It's a new link, so let's save it
	err = u.db.InsertUrl(ctx, url)
	if err != nil {
		return nil, 0,  err
	}

	count, err := u.db.GetCountShortUrls(ctx)
	if err != nil {
		return nil, 0,  err
	}

	return &url, count, nil
}

func (u *URLShortener) GetOriginalURL(ctx context.Context, shortUrl string) (string, error) {

	//use trategy cashe aside
	//first check in redis
	redisUrl, err := u.cache.Get(ctx, shortUrl)
	if err == nil {
		return fmt.Sprintf("%v", redisUrl), nil
	}
	//if cache miss, query the database
	url, err := u.db.GetShortUrl(ctx, shortUrl)
	if err != nil {
		return "", err
	}

	//store in the redis
	err = u.cache.Set(ctx, shortUrl, url, time.Hour)
	if err != nil {
		u.logger.Error("redis insertion error", slog.String("message", err.Error()))
	}

	return url.LongURL, nil
}


func (u *URLShortener) DeleteShortUrl(ctx context.Context, shortUrl string) (error) {

	 err := u.db.DeleteShortUrl(ctx, shortUrl)
	if err != nil {
		return  err
	}

	return nil
}
