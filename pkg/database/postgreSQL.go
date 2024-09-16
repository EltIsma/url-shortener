package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	db *pgxpool.Pool
}

func NewPG(connURL string) (*Postgres, error) {
	config, err := pgxpool.ParseConfig(connURL)
	if err != nil {
		return nil, fmt.Errorf(": %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf(": %w", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf(": %w", err)
	}

	return &Postgres{
		db: pool,
	}, nil
}

func(pg *Postgres) GetConn() *pgxpool.Pool{
	return pg.db
}


func (pg *Postgres) Close() {
	pg.db.Close()
}
