package local

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"url-shortener/internal/domain"
)

type repository struct {
	Long map[string]struct{}
	Short map[string]domain.URL
	mu        sync.RWMutex
}

func New() *repository {
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	return &repository{
		Long: make(map[string]struct{}),
		Short: make(map[string]domain.URL),
		mu:        sync.RWMutex{}}
}

func (r *repository) InsertUrl(ctx context.Context, url domain.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Long[url.LongURL] = struct{}{}
	r.Short[url.ShortURL] = url
	return nil
}

func (r *repository) GetShortUrl(ctx context.Context, url string) (*domain.URL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.Short[url]; !ok {
		return nil, domain.ErrOriginalURLNotFound
	}
	res := r.Short[url]

	return &res, nil
}

func (r *repository) GetCountShortUrls(ctx context.Context) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.Short), nil
}

func (r *repository) DeleteShortUrl(ctx context.Context, shortURL string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.Short[shortURL]; !ok {
		return errors.New("not found short url")
	}
	delete(r.Long, r.Short[shortURL].LongURL)
	delete(r.Short, shortURL)
	return nil
  }

  func (r *repository) GetByLongUrl(ctx context.Context, url string) (*domain.URL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.Long[url]; !ok {
		return nil, domain.ErrOriginalURLNotFound
	}
	res := r.Short[url]

	return &res, nil
}

