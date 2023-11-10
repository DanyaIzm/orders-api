package redisstorage

import (
	"context"
	"os"

	"github.com/danyaizm/orders-api/storage"
	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	orderRepo *OrderRepo
	clinet    *redis.Client
}

func NewRedisStorage(ctx context.Context) (*RedisStorage, error) {
	addr, exists := os.LookupEnv("REDIS_ADDR")

	if !exists {
		addr = "localhost:6379"
	}

	client := redis.NewClient(
		&redis.Options{
			Addr: addr,
		},
	)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	s := &RedisStorage{
		clinet: client,
		orderRepo: &OrderRepo{
			client: client,
		},
	}

	return s, nil
}

func (s *RedisStorage) Close() {
	s.clinet.Close()
}

func (s *RedisStorage) OrderRepo() storage.OrderRepo {
	return s.orderRepo
}
