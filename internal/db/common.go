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

// OrderDao defines the persistence operations required to persist orders in storage.
// Implementations are responsible for validation and writing orders (ID and items) to the DB.
type OrderDao interface {
	CreateOrder(context.Context, Order) error
}

type Product struct {
	Id       ID      `json:"id" validate:"required"`
	Name     string  `json:"name" validate:"required"`
	Price    float32 `json:"price" validate:"gte=0"`
	Category string  `json:"category" validate:"required"`
}

// ProductDao defines read operations for product data used by services.
// Implementations provide GetAllProducts and GetProduct by ID.
type ProductDao interface {
	// Define Product DAO methods here.
	GetProduct(context.Context, ID) (Product, error)
	GetAllProducts(context.Context) ([]Product, error)
}

// SearchResult represents the asynchronous result of searching coupon files.
// Validate waits for search completion and returns whether the coupon exists.
type SearchResult interface {
	Validate() (bool, error)
}

// CouponDao encapsulates searching for promo codes in configured files.
// The SearchForCouponInGivenFiles returns a SearchResult for async validation.
type CouponDao interface {
	SearchForCouponInGivenFiles(openapi.OrderReq) (SearchResult, error)
}
