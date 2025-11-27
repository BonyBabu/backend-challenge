package services

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	db "backend-challenge/internal/db"
	dbmocks "backend-challenge/internal/db/mocks"
	openapi "backend-challenge/internal/generated/openapi"
)

func TestGetProduct(t *testing.T) {
	tests := []struct {
		name      string
		productID int64
		product   db.Product
		prodErr   error
		wantCode  int
	}{
		{name: "product not exist", productID: int64(999), product: db.Product{}, prodErr: db.ErrNotFound(sql.ErrNoRows), wantCode: http.StatusNotFound},
		{name: "product exists", productID: int64(1), product: db.Product{Id: "1", Name: "Product One", Price: 100, Category: "cat"}, prodErr: nil, wantCode: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			pd := dbmocks.NewMockProductDao(ctrl)
			pd.EXPECT().GetProduct(gomock.Any(), db.ID(strconv.FormatInt(tt.productID, 10))).Return(tt.product, tt.prodErr).Times(1)
			svc := NewProductAPIService(pd)
			res, err := svc.GetProduct(context.Background(), tt.productID)
			if err != nil {
				t.Fatalf("Service error: %v", err)
			}
			assert.Equal(t, tt.wantCode, res.Code)
		})
	}
}

func TestListProducts(t *testing.T) {
	tests := []struct {
		name      string
		products  []db.Product
		prodErr   error
		wantCode  int
		wantCount int
	}{
		{name: "list all products", products: []db.Product{{Id: "1", Name: "p1", Price: 100, Category: "c"}, {Id: "2", Name: "p2", Price: 200, Category: "c"}}, prodErr: nil, wantCode: http.StatusOK, wantCount: 2},
		{name: "no products", products: []db.Product{}, prodErr: nil, wantCode: http.StatusOK, wantCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			pd := dbmocks.NewMockProductDao(ctrl)
			pd.EXPECT().GetAllProducts(gomock.Any()).Return(tt.products, tt.prodErr).Times(1)
			svc := NewProductAPIService(pd)
			res, err := svc.ListProducts(context.Background())
			if err != nil {
				t.Fatalf("Service error: %v", err)
			}
			assert.Equal(t, tt.wantCode, res.Code)
			if tt.wantCode == http.StatusOK {
				if body, ok := res.Body.([]openapi.Product); ok {
					assert.Equal(t, tt.wantCount, len(body))
				} else {
					t.Fatalf("unexpected body type: %T", res.Body)
				}
			}
		})
	}
}
