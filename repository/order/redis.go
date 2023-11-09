package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	model "github.com/danyaizm/orders-api/models"
	"github.com/danyaizm/orders-api/repository"
	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client *redis.Client
}

func getOrderIDKey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}

func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode order: %w", err)
	}

	key := getOrderIDKey(order.ID)

	txn := r.Client.TxPipeline()

	res := txn.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to insert order: %w", err)
	}

	if err := txn.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add orders set: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return nil
}

func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (*model.Order, error) {
	key := getOrderIDKey(id)

	value, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return &model.Order{}, repository.ErrorNotExist
	} else if err != nil {
		return &model.Order{}, fmt.Errorf("failed to fetch from redis by id: %w", err)
	}

	var order *model.Order
	if err := json.Unmarshal([]byte(value), &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshall value: %w", err)
	}

	return order, nil
}

func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	key := getOrderIDKey(id)

	txn := r.Client.TxPipeline()

	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return repository.ErrorNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("failed to delete: %w", err)
	}

	if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to remove item from orders set: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return nil
}

func (r *RedisRepo) Update(ctx context.Context, order *model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}

	key := getOrderIDKey(order.ID)

	if err := r.Client.SetXX(ctx, key, string(data), 0).Err(); err != nil {
		return err
	}

	return nil
}

func (r *RedisRepo) FindAll(ctx context.Context, page repository.FindAllPage) (*repository.FindResult, error) {
	res := r.Client.SScan(ctx, "orders", uint64(page.Offset), "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return &repository.FindResult{}, fmt.Errorf("failed to fetch all order keys from set: %w", err)
	}

	if len(keys) == 0 {
		return &repository.FindResult{
			Orders: []model.Order{},
			Cursor: cursor,
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return &repository.FindResult{}, fmt.Errorf("failed to get orders: %w", err)
	}

	orders := make([]model.Order, len(xs))

	for i, x := range xs {
		x := x.([]byte)

		var order model.Order
		if err := json.Unmarshal(x, &order); err != nil {
			return &repository.FindResult{}, fmt.Errorf("failed to unmarshal one of the results: %w", err)
		}

		orders[i] = order
	}

	return &repository.FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}
