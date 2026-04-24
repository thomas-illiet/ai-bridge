package services

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const reloadChannel = "aibridge:reload"

// ReloadPublisher signals the bridge service to reload its providers.
type ReloadPublisher interface {
	PublishReload(ctx context.Context) error
}

// RedisReloadPublisher publishes reload signals via Redis pub/sub.
type RedisReloadPublisher struct {
	rdb *redis.Client
}

func NewRedisReloadPublisher(rdb *redis.Client) *RedisReloadPublisher {
	return &RedisReloadPublisher{rdb: rdb}
}

func (p *RedisReloadPublisher) PublishReload(ctx context.Context) error {
	return p.rdb.Publish(ctx, reloadChannel, "reload").Err()
}

// ReloadChannel is the Redis pub/sub channel name used for reload signals.
func ReloadChannel() string { return reloadChannel }
