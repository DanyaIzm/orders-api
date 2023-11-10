package storage

import (
	"context"

	model "github.com/danyaizm/orders-api/models"
)

type OrderRepo interface {
	Insert(context.Context, model.Order) error
	FindByID(context.Context, uint64) (*model.Order, error)
	DeleteByID(context.Context, uint64) error
	Update(context.Context, *model.Order) error
	FindAll(context.Context, FindAllPage) (*FindResult, error)
}
