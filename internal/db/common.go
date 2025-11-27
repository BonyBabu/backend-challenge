//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -destination=mocks/mock_db.go -package=mocks backend-challenge/internal/db OrderDao,ProductDao,CouponDao,SearchResult
package db

// NOTE: To regenerate mocks, run `go generate ./...` or `go generate` from this package.

import (
	"backend-challenge/internal/generated/openapi"
	"context"
)

type ID = string
type ErrNotFound error

type Item struct {
	ProductID ID    `json:"product_id" validate:"required"`
	Quantity  int32 `json:"quantity" validate:"gt=0"`
}

type Order struct {
	ID    ID     `json:"id" validate:"required"`
	Items []Item `json:"items" validate:"required,dive"`
}

type OrderDao interface {
	CreateOrder(context.Context, Order) error
}

type Product struct {
	Id       ID      `json:"id" validate:"required"`
	Name     string  `json:"name" validate:"required"`
	Price    float32 `json:"price" validate:"gte=0"`
	Category string  `json:"category" validate:"required"`
}

type ProductDao interface {
	// Define Product DAO methods here.
	GetProduct(context.Context, ID) (Product, error)
	GetAllProducts(context.Context) ([]Product, error)
}

type SearchResult interface {
	Validate() (bool, error)
}

type CouponDao interface {
	SearchForCouponInGivenFiles(openapi.OrderReq) (SearchResult, error)
}
