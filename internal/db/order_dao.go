package db

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/go-playground/validator/v10"
)

type OrderDaoImpl struct {
	db *sql.DB
}

var _ OrderDao = &OrderDaoImpl{}

func NewOrderDao(db *sql.DB) OrderDao {
	return &OrderDaoImpl{db: db}
}

func (generalOrder *OrderDaoImpl) CreateOrder(ctx context.Context, order Order) error {
	// Validate order structure
	validate := validator.New()
	if err := validate.Struct(order); err != nil {
		return err
	}
	query := "INSERT INTO orders (id, items) VALUES (?, ?)"
	items, err := json.Marshal(order.Items)
	if err != nil {
		return err
	}
	if _, err := generalOrder.db.ExecContext(ctx, query, order.ID, items); err != nil {
		return err
	}
	return nil
}
