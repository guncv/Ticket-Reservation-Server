package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/infra/log"
	"github.com/guncv/ticket-reservation-server/internal/shared"
)

type RedisClient interface {
	Set(ctx context.Context, payload RedisPayload) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, keys ...string) error
	FlushAll(ctx context.Context) error
}

type redisClient struct {
	client *redis.Client
	cfg    *config.Config
	log    log.Logger
}

type RedisPayload struct {
	Key   string
	Value interface{}
	TTL   time.Duration
}

func NewRedisClient(cfg *config.Config, logger log.Logger) RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisConfig.Host, cfg.RedisConfig.Port),
		Password: cfg.RedisConfig.Password,
		DB:       cfg.RedisConfig.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), shared.TestTimeout)
	defer cancel()

	if err := rdb.Ping().Err(); err != nil {
		logger.Error(ctx, "Failed to connect to Redis", err)
	}

	return &redisClient{
		client: rdb,
		cfg:    cfg,
		log:    logger,
	}
}

func (r *redisClient) Set(ctx context.Context, payload RedisPayload) error {
	if err := r.client.Set(payload.Key, payload.Value, payload.TTL).Err(); err != nil {
		r.log.Error(ctx, "[Redis Client: Set] Error", err)
		return err
	}

	return nil
}

func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	res, err := r.client.Get(key).Result()
	if err != nil {
		r.log.Error(ctx, "[Redis Client: Get] Error", err)
		return "", err
	}

	return res, nil
}

func (r *redisClient) Delete(ctx context.Context, keys ...string) error {
	if err := r.client.Del(keys...).Err(); err != nil {
		r.log.Error(ctx, "[Redis Client: Delete] Error", err)
		return err
	}

	return nil
}

func (r *redisClient) FlushAll(ctx context.Context) error {
	if err := r.client.FlushAll().Err(); err != nil {
		r.log.Error(ctx, "[Redis Client: FlushAll] Error", err)
		return err
	}

	return nil
}
