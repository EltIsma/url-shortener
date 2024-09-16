package httpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"url-shortener/internal/domain"
	"url-shortener/internal/ports/httpServer/request"
	"url-shortener/internal/ports/httpServer/response"
	"url-shortener/pkg/metrics"
)


type URLShortenerService interface {
	Create(ctx context.Context, url string) (*domain.URL, int, error)
	GetOriginalURL(ctx context.Context, shortUrl string) (string, error)
	DeleteShortUrl(ctx context.Context, shortUrl string) (error) 
}

type EncoderService interface {
	Base62Encode(number uint64) string
}

type RepresenrService interface {
	Home(http.ResponseWriter)
}

type Handler struct {
	urlshortener URLShortenerService
	logger       *slog.Logger
	render       RepresenrService
	metrics      *metrics.PrometheusMetrics
}

func NewHandler(logger *slog.Logger, urlshortener URLShortenerService, render RepresenrService, metrics *metrics.PrometheusMetrics) *Handler {
	return &Handler{
		logger:       logger,
		urlshortener: urlshortener,
		render:       render,
		metrics:      metrics,
	}
}

func (h *Handler) Homepage(w http.ResponseWriter, _ *http.Request) {
	h.render.Home(w)
}

func (h *Handler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var input request.UrlRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ResultJSON(w, http.StatusBadRequest, map[string]any{"message": err.Error()})
		return
	}

	newUrl, count, err := h.urlshortener.Create(r.Context(), input.URL)
	if err != nil {
		h.logger.Error("failed to create short url", slog.String("error", err.Error()))
		response.ResultJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}

	h.metrics.UrlsTotal.Set(float64(count))
	body := map[string]any{
		"short_url":    newUrl.ShortURL,
		"original_url": newUrl.LongURL,
	}
	h.metrics.SuccessRequest.Inc()
	response.ResultJSON(w, http.StatusOK, body)

}

func (h *Handler) RedirectionToUrl(w http.ResponseWriter, r *http.Request) {
	shortUrl := r.PathValue("shortUrl")
	original_url, err := h.urlshortener.GetOriginalURL(r.Context(), shortUrl)
	if err != nil {
		h.logger.Error("failed to get url", slog.String("error", err.Error()))
		response.ResultJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
	}
	h.metrics.RedirectsTotal.Inc()
	h.metrics.Redirects.WithLabelValues(original_url).Inc()
	http.Redirect(w, r, original_url, http.StatusMovedPermanently)

}


func (h *Handler) DeleteShortURL(w http.ResponseWriter, r *http.Request) {
	var input request.UrlRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ResultJSON(w, http.StatusBadRequest, map[string]any{"message": err.Error()})
		return
	}

	 err := h.urlshortener.DeleteShortUrl(r.Context(), input.URL)
	if err != nil {
		h.logger.Error("failed to create short url", slog.String("error", err.Error()))
		response.ResultJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}

	body := map[string]any{
		"result of deleting":   "ok",
	}
	h.metrics.SuccessRequest.Inc()
	response.ResultJSON(w, http.StatusOK, body)

}