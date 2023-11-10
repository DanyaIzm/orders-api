package storage

import model "github.com/danyaizm/orders-api/models"

type FindAllPage struct {
	Size   uint64
	Offset uint64
}

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}
