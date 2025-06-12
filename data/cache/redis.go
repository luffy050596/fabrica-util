// Package cache provides Redis cache implementation with connection management
package cache

import (
	"context"
	"log/slog"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

// NewRedis creates a new Redis client and returns a Cacheable interface
// It also returns a cleanup function to close the connection
func NewRedis(c *redis.Options) (rdb redis.UniversalClient, cleanup func(), err error) {
	rdb = redis.NewClient(c)

	cleanup = func() {
		if err0 := rdb.(*redis.Client).Close(); err0 != nil {
			slog.Error("redis close failed", "error", err0)
		} else {
			slog.Info("redis close success")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.DialTimeout)
	defer cancel()

	if err = rdb.Ping(ctx).Err(); err != nil {
		err = errors.Wrapf(err, "redis ping failed")
		return
	}

	return
}

// NewRedisCluster creates a new Redis cluster client and returns a Cacheable interface
// It also returns a cleanup function to close the connection
func NewRedisCluster(c *redis.ClusterOptions) (rdb redis.UniversalClient, cleanup func(), err error) {
	rdb = redis.NewClusterClient(c)

	cleanup = func() {
		if err0 := rdb.(*redis.ClusterClient).Close(); err0 != nil {
			slog.Error("redis cluster close failed", "error", err0)
		} else {
			slog.Info("redis cluster close success")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.DialTimeout)
	defer cancel()

	if err = rdb.Ping(ctx).Err(); err != nil {
		err = errors.Wrapf(err, "redis cluster ping failed")
		return
	}

	return
}
