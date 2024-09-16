package httpserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/domain"
	urlMocks "url-shortener/internal/mocks"
	"url-shortener/internal/ports/httpServer/request"
	"url-shortener/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_CreateShortURL(t *testing.T) {
	t.Run("Successful creation", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		m := metrics.NewMetrics(reg)
		logger := &slog.Logger{}
		urlshortener := urlMocks.NewURLShortenerService(t)
		render := urlMocks.NewRepresenrService(t)
		handler := NewHandler(logger, urlshortener, render, m)

		input := request.UrlRequest{URL: "https://example.com"}
		jsonInput, _ := json.Marshal(input)

		newURL := &domain.URL{ShortURL: "shortURL", LongURL: "https://example.com"}
		urlshortener.On("Create", mock.Anything, "https://example.com").Return(newURL, 10, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/data/shorten", bytes.NewReader(jsonInput))
		rr := httptest.NewRecorder()

		handler.CreateShortURL(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var body map[string]any
		json.Unmarshal(rr.Body.Bytes(), &body)

		assert.Equal(t, "shortURL", body["short_url"])
		assert.Equal(t, "https://example.com", body["original_url"])
		urlshortener.AssertExpectations(t)
	})

	t.Run("Invalid JSON input", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		m := metrics.NewMetrics(reg)
		logger := &slog.Logger{}
		urlshortener := urlMocks.NewURLShortenerService(t)
		render := urlMocks.NewRepresenrService(t)
		handler := NewHandler(logger, urlshortener, render, m)

		req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewReader([]byte("invalid json")))
		rr := httptest.NewRecorder()

		handler.CreateShortURL(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var body map[string]any
		json.Unmarshal(rr.Body.Bytes(), &body)

		assert.Equal(t, "invalid character 'i' looking for beginning of value", body["message"])
	})

	t.Run("Error creating short URL", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		m := metrics.NewMetrics(reg)
		opts := &slog.HandlerOptions{}
		logHandler := slog.NewJSONHandler(os.Stdout, opts)
		logger := slog.New(logHandler)
		urlshortener := urlMocks.NewURLShortenerService(t)
		render := urlMocks.NewRepresenrService(t)
		handler := NewHandler(logger, urlshortener, render, m)

		input := request.UrlRequest{URL: "https://example.com"}
		jsonInput, _ := json.Marshal(input)

		urlshortener.On("Create", mock.Anything, "https://example.com").Return(nil, 0, errors.New("database error"))

		req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewReader(jsonInput))
		rr := httptest.NewRecorder()

		handler.CreateShortURL(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var body map[string]any
		json.Unmarshal(rr.Body.Bytes(), &body)

		assert.Equal(t, "database error", body["message"])
		urlshortener.AssertExpectations(t)
	})
}

func TestHandler_RedirectionToUrl(t *testing.T) {
	t.Run("Successful redirection", func(t *testing.T) {
		logger := &slog.Logger{}
		urlshortener := urlMocks.NewURLShortenerService(t)
		render := urlMocks.NewRepresenrService(t)
		reg := prometheus.NewRegistry()
		m := metrics.NewMetrics(reg)
		handler := NewHandler(logger, urlshortener, render, m)

		shortURL := "shortURL"
		originalURL := "https://example.com"

		urlshortener.On("GetOriginalURL", mock.Anything, "").Return(originalURL, nil)

		req := httptest.NewRequest(http.MethodGet, "/"+shortURL, nil)
		rr := httptest.NewRecorder()

		handler.RedirectionToUrl(rr, req)

		assert.Equal(t, http.StatusMovedPermanently, rr.Code)
		assert.Equal(t, originalURL, rr.Header().Get("Location"))
		urlshortener.AssertExpectations(t)
	})

	t.Run("Error getting original URL", func(t *testing.T) {
		opts := &slog.HandlerOptions{}
		logHandler := slog.NewJSONHandler(os.Stdout, opts)
		logger := slog.New(logHandler)
		urlshortener := urlMocks.NewURLShortenerService(t)
		render := urlMocks.NewRepresenrService(t)
		reg := prometheus.NewRegistry()
		m := metrics.NewMetrics(reg)
		handler := NewHandler(logger, urlshortener, render, m)

		shortURL := "shortURL"

		urlshortener.On("GetOriginalURL", mock.Anything, "").Return("", errors.New("database error"))

		req := httptest.NewRequest(http.MethodGet, "/"+shortURL, nil)
		rr := httptest.NewRecorder()

		handler.RedirectionToUrl(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var body map[string]any
		json.Unmarshal(rr.Body.Bytes(), &body)

		assert.Equal(t, "database error", body["message"])
		urlshortener.AssertExpectations(t)
		//assert.Equal(t, float64(0), metrics.RedirectsTotal.Get())
	})
}
