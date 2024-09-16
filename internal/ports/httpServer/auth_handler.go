package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"url-shortener/internal/domain"

	"github.com/go-playground/validator/v10"
)

type ServiceAuth interface {
	Register(ctx context.Context, nickname, password string) error
	Login(ctx context.Context, nickname, password string) (*domain.Tokens, *domain.User, error)
	Refresh(ctx context.Context, token string) (*domain.Tokens, error)
}

type AuthHandler struct {
	logger *slog.Logger
	auth   ServiceAuth
}

func NewAuthHandler(logger *slog.Logger, auth ServiceAuth) *AuthHandler {
	return &AuthHandler{
		logger: logger,
		auth:   auth,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	buf, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		ProcessError(w, "can not read request body", http.StatusBadRequest)
		return
	}

	var register registerRequest
	err = json.Unmarshal(buf, &register)
	if err != nil {
		ProcessError(w, "can not unmarshal request body", http.StatusBadRequest)
		return
	}

	if err = validator.New().Struct(register); err != nil {
		var validateErrs validator.ValidationErrors
		errors.As(err, &validateErrs)

		ProcessError(w, ValidationError(validateErrs), http.StatusBadRequest)
		return
	}

	err = h.auth.Register(r.Context(), register.Nickname, register.Password)
	if err != nil {
		if errors.Is(err, domain.ErrNicknameAlreadyExist) {
			ProcessError(w, domain.ErrNicknameAlreadyExist.Error(), http.StatusBadRequest)
			return
		}

		h.logger.Error("failed to register user", slog.String("error", err.Error()))
		ProcessError(w, "failed to register user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	buf, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		ProcessError(w, "can not read request body", http.StatusBadRequest)
		return
	}

	var register loginRequest
	err = json.Unmarshal(buf, &register)
	if err != nil {
		ProcessError(w, "can not unmarshal request body", http.StatusBadRequest)
		return
	}

	if err = validator.New().Struct(register); err != nil {
		var validateErrs validator.ValidationErrors
		errors.As(err, &validateErrs)

		ProcessError(w, ValidationError(validateErrs), http.StatusBadRequest)
		return
	}

	tokens, user, err := h.auth.Login(r.Context(), register.Nickname, register.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			ProcessError(w, domain.ErrInvalidCredentials.Error(), http.StatusUnauthorized)
			return
		}

		h.logger.Error("failed to login user", slog.String("error", err.Error()))
		ProcessError(w, "failed to login user", http.StatusInternalServerError)
		return
	}

	tokenResp := tokenResponse{
		UserID:       user.ID,
		Nickname:     user.Nickname,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

	payload, err := json.Marshal(tokenResp)
	if err != nil {
		h.logger.Error("can not marshal response body", slog.String("error", err.Error()))
		ProcessError(w, "can not marshal response body", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(payload)
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	buf, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		ProcessError(w, "can not read request body", http.StatusBadRequest)
		return
	}

	var refresh refreshRequest
	err = json.Unmarshal(buf, &refresh)
	if err != nil {
		ProcessError(w, "can not unmarshal request body", http.StatusBadRequest)
		return
	}

	if err = validator.New().Struct(refresh); err != nil {
		var validateErrs validator.ValidationErrors
		errors.As(err, &validateErrs)

		ProcessError(w, ValidationError(validateErrs), http.StatusBadRequest)
		return
	}

	tokens, err := h.auth.Refresh(r.Context(), refresh.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			ProcessError(w, "token not found", http.StatusUnauthorized)
			return
		}

		h.logger.Error("failed to refresh tokens", slog.String("error", err.Error()))
		ProcessError(w, "failed to refresh tokens", http.StatusInternalServerError)
		return
	}

	tokenResp := tokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

	payload, err := json.Marshal(tokenResp)
	if err != nil {
		h.logger.Error("can not marshal response body", slog.String("error", err.Error()))
		ProcessError(w, "can not marshal response body", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(payload)
	w.WriteHeader(http.StatusOK)
}
