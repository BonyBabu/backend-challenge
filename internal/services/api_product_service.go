package services

import (
	"backend-challenge/internal/db"
	openapi "backend-challenge/internal/generated/openapi"
	"context"
	"net/http"
	"strconv"
)

// ProductAPIService implements business logic for the ProductAPI defined by the generated OpenAPI.
// It exposes product listing and lookup functionality using a `db.ProductDao`.
// Methods are wired in generated API router as an implementation of ProductAPIServicer.
type ProductAPIService struct {
	productDao db.ProductDao
}

// NewProductAPIService creates a default api service
func NewProductAPIService(productDao db.ProductDao) *ProductAPIService {
	return &ProductAPIService{productDao: productDao}
}

// ListProducts - List products
func (s *ProductAPIService) ListProducts(ctx context.Context) (openapi.ImplResponse, error) {
	products, err := s.productDao.GetAllProducts(ctx)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, nil), err
	}
	openapiProducts := make([]openapi.Product, 0, len(products))
	for _, p := range products {
		openapiProduct := openapi.Product{
			Id:       p.Id,
			Name:     p.Name,
			Price:    p.Price,
			Category: p.Category,
		}
		openapiProducts = append(openapiProducts, openapiProduct)
	}
	return openapi.Response(http.StatusOK, openapiProducts), nil
}

// GetProduct - Find product by ID
func (s *ProductAPIService) GetProduct(ctx context.Context, productId int64) (openapi.ImplResponse, error) {
	productIdStr := strconv.FormatInt(productId, 10)
	product, err := s.productDao.GetProduct(ctx, db.ID(productIdStr))
	if err != nil {
		if _, ok := err.(db.ErrNotFound); ok {
			return openapi.Response(http.StatusNotFound, nil), nil
		}
		return openapi.Response(http.StatusInternalServerError, nil), err
	}
	openapiProduct := openapi.Product{
		Id:       product.Id,
		Name:     product.Name,
		Price:    product.Price,
		Category: product.Category,
	}
	return openapi.Response(http.StatusOK, openapiProduct), nil
}
