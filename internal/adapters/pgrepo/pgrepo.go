package pgrepo

import (
	"context"
	"errors"
	"fmt"
	"url-shortener/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryPG struct {
	conn *pgxpool.Pool
}

func NewRepositoruPG(conn *pgxpool.Pool) *RepositoryPG {
	return &RepositoryPG{
		conn: conn,
	}
}

func (pg *RepositoryPG) InsertUrl(ctx context.Context, url domain.URL) error {
	_, err := pg.conn.Exec(ctx, "INSERT INTO short_urls (unique_id, short_url, long_url) VALUES($1, $2, $3)", url.Id, url.ShortURL, url.LongURL)
	if err != nil {
		return err
	}

	return nil
}

func (pg *RepositoryPG) GetByLongUrl(ctx context.Context, url string) (*domain.URL, error) {
	var link domain.URL
	err := pg.conn.QueryRow(ctx, "SELECT unique_id, short_url, long_url FROM short_urls WHERE long_url = $1", url).Scan(&link.Id, &link.ShortURL, &link.LongURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOriginalURLNotFound
		}
		return nil, err
	}

	return &link, nil
}

func (pg *RepositoryPG) GetShortUrl(ctx context.Context, url string) (*domain.URL, error) {
	var link domain.URL
	err := pg.conn.QueryRow(ctx, "SELECT unique_id, short_url, long_url FROM short_urls WHERE short_url = $1", url).Scan(&link.Id, &link.ShortURL, &link.LongURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOriginalURLNotFound
		}
		return nil, err
	}

	return &link, nil
}

func (pg *RepositoryPG) GetCountShortUrls(ctx context.Context) (int, error) {
	var count int
	err := pg.conn.QueryRow(ctx, "SELECT COUNT(short_url) FROM short_urls").Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (pg *RepositoryPG) DeleteShortUrl(ctx context.Context, shortURL string) error {
	_, err := pg.conn.Exec(ctx, "DELETE FROM short_urls WHERE short_url = $1", shortURL)
	if err != nil {
	  return err
	}
	return nil
  }

func (pg *RepositoryPG) SaveUser(ctx context.Context, user *domain.User) (string, error) {
	row := pg.conn.QueryRow(ctx, "INSERT INTO users(nickname, password_hash) VALUES ($1, $2) RETURNING id", user.Nickname, user.PasswordHash)

	var id string
	err := row.Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName != "" {
			return "", domain.ErrNicknameAlreadyExist
		}

		return "", fmt.Errorf("storage.pg.SaveUser: %w", err)
	}

	return id, nil
}

func (pg *RepositoryPG) GetUser(ctx context.Context, nickname string) (*domain.User, error) {
	row := pg.conn.QueryRow(ctx, "SELECT id, nickname, password_hash FROM users WHERE nickname = $1", nickname)

	var user domain.User
	err := row.Scan(&user.ID, &user.Nickname, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("storage.pg.GetUser: %w", err)
	}

	return &user, nil
}

func (pg *RepositoryPG) SetSession(ctx context.Context, userID string, session *domain.Session) error {
	_, err := pg.conn.Exec(ctx, "UPDATE users SET refresh_token = $1, expires_at = $2 WHERE id = $3", session.RefreshToken, session.ExpiresAt, userID)
	if err != nil {
		return fmt.Errorf("storage.pg.SetSession: %w", err)
	}

	return nil
}

func (pg *RepositoryPG) GetBySession(ctx context.Context, refreshToken string) (*domain.User, error) {
	row := pg.conn.QueryRow(ctx, "SELECT id, nickname FROM users WHERE refresh_token = $1", refreshToken)

	var user domain.User
	err := row.Scan(&user.ID, &user.Nickname)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("storage.pg.GetBySession: %w", err)
	}

	return &user, nil
}
