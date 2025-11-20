package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dmehra2102/budget-tracker/internal/config"
	"github.com/redis/go-redis/v9"
)

type CacheService interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
}

type cacheService struct {
	client *redis.Client
	cfg    *config.Config
}

func NewCacheService(client *redis.Client, cfg *config.Config) CacheService {
	return &cacheService{
		client: client,
		cfg:    cfg,
	}
}

func (s *cacheService) Get(ctx context.Context, key string, dest any) error {
	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

func (s *cacheService) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if ttl == 0 {
		ttl = s.cfg.Redis.TTL
	}

	return s.client.Set(ctx, key, data, ttl).Err()
}

func (s *cacheService) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

func (s *cacheService) DeletePattern(ctx context.Context, pattern string) error {
	iter := s.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := s.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
