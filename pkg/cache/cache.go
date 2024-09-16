package cache

import (
	"context"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key any, value any, ttl time.Duration) error
	Get(ctx context.Context, key any) (value any, err error)
}
