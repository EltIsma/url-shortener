package ratelimiter

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis_rate/v9"
)

var ErrRateLimited = errors.New("rate limited")

var Limiter *redis_rate.Limiter

func RateLimit(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			res, err := Limiter.Allow(r.Context(), "url-shortener", redis_rate.PerMinute(10))
			if err != nil {
				logger.Error("Rate limiter", slog.String("error", err.Error()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			h := w.Header()
			h.Set("RateLimit-Remaining", strconv.Itoa(res.Remaining))

			if res.Allowed == 0 {
				// We are rate limited.

				seconds := int(res.RetryAfter / time.Second)
				h.Set("RateLimit-RetryAfter", strconv.Itoa(seconds))

				// Stop processing and return the error.
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			// Continue processing as normal.
			next.ServeHTTP(w, r)
		})
	}
}
