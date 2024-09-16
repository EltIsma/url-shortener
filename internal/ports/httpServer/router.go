package httpserver

import (
	"log/slog"
	"net/http"
	"url-shortener/pkg/jwt"
	ratelimiter "url-shortener/pkg/rate-limiter/leaking_bucket"

	"github.com/go-redis/redis_rate/v9"
)

func InitRouter(handler *Handler, auth *AuthHandler, logger *slog.Logger, rL *redis_rate.Limiter, manager jwt.TokenManager) http.Handler {
	ratelimiter.Limiter = rL
	rateLimiter := ratelimiter.RateLimit(logger)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /user/register", auth.Register)
	mux.HandleFunc("POST /user/login", auth.Login)
	mux.HandleFunc("POST /user/refresh", auth.RefreshTokens)

	authMiddleware := jwt.Validate(manager)
	mux.Handle("DELETE /api/v1/data/shorten/delete", authMiddleware(http.HandlerFunc(handler.DeleteShortURL)))

	mux.HandleFunc("POST /api/v1/data/shorten", handler.CreateShortURL)
	mux.HandleFunc("GET /api/v1/{shortUrl}", handler.RedirectionToUrl)
	mux.HandleFunc("GET /{shortUrl}", handler.RedirectionToUrl)
	mux.HandleFunc("GET /", handler.Homepage)
	muxWithLimiter := rateLimiter(mux)
	return muxWithLimiter
}
