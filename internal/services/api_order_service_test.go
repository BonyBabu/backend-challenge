package services

import (
	"context"
	"errors"
	"net/http"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	db "backend-challenge/internal/db"
	dbmocks "backend-challenge/internal/db/mocks"
	openapi "backend-challenge/internal/generated/openapi"
)

// testCouponDao is a simple test implementation of db.CouponDao used to avoid mocking the
// file scanning logic; it returns search results based on the `found` field.
type testCouponDao struct {
	found bool
}

func (t *testCouponDao) SearchForCouponInGivenFiles(ctx context.Context, orderReq openapi.OrderReq) (db.SearchResult, error) {
	return &testSearchResult{found: t.found}, nil
}

type testSearchResult struct {
	found bool
}

func (t *testSearchResult) Validate() (bool, error) { return t.found, nil }

func TestPlaceOrder(t *testing.T) {
	type args struct {
		req openapi.OrderReq
	}
	tests := []struct {
		name       string
		args       args
		setupMocks func(ctrl *gomock.Controller) (db.OrderDao, db.ProductDao, db.CouponDao)
		wantCode   int
	}{
		{
			name: "invalid coupon length",
			args: args{req: openapi.OrderReq{CouponCode: "SHORT", Items: []openapi.OrderReqItemsInner{{ProductId: "1", Quantity: 1}}}},
			setupMocks: func(ctrl *gomock.Controller) (db.OrderDao, db.ProductDao, db.CouponDao) {
				oc := dbmocks.NewMockOrderDao(ctrl)
				pc := dbmocks.NewMockProductDao(ctrl)
				return oc, pc, &testCouponDao{found: false}
			},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid quantity",
			args: args{req: openapi.OrderReq{CouponCode: "HAPPYHOURS", Items: []openapi.OrderReqItemsInner{{ProductId: "1", Quantity: 0}}}},
			setupMocks: func(ctrl *gomock.Controller) (db.OrderDao, db.ProductDao, db.CouponDao) {
				oc := dbmocks.NewMockOrderDao(ctrl)
				pc := dbmocks.NewMockProductDao(ctrl)
				return oc, pc, &testCouponDao{found: true}
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "invalid product specified",
			args: args{req: openapi.OrderReq{CouponCode: "HAPPYHOURS", Items: []openapi.OrderReqItemsInner{{ProductId: "invalid-prod", Quantity: 1}}}},
			setupMocks: func(ctrl *gomock.Controller) (db.OrderDao, db.ProductDao, db.CouponDao) {
				oc := dbmocks.NewMockOrderDao(ctrl)
				pc := dbmocks.NewMockProductDao(ctrl)
				// mock product failure
				pc.EXPECT().GetProduct(gomock.Any(), db.ID("invalid-prod")).Return(db.Product{}, errors.New("not found"))
				return oc, pc, &testCouponDao{found: true}
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "create test order successfully",
			args: args{req: openapi.OrderReq{CouponCode: "HAPPYHOURS", Items: []openapi.OrderReqItemsInner{{ProductId: "1", Quantity: 2}}}},
			setupMocks: func(ctrl *gomock.Controller) (db.OrderDao, db.ProductDao, db.CouponDao) {
				oc := dbmocks.NewMockOrderDao(ctrl)
				pc := dbmocks.NewMockProductDao(ctrl)
				// mock product lookup
				pc.EXPECT().GetProduct(gomock.Any(), db.ID("1")).Return(db.Product{Id: "1", Name: "Product 1", Price: 100, Category: "cat"}, nil)
				// expect CreateOrder to be called and succeed
				oc.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).Return(nil)
				return oc, pc, &testCouponDao{found: true}
			},
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			orderDao, productDao, couponDao := tt.setupMocks(ctrl)
			svc := NewOrderAPIServiceWithCouponDao(orderDao, productDao, couponDao)
			res, err := svc.PlaceOrder(context.Background(), tt.args.req)
			if err != nil {
				t.Fatalf("Service error: %v", err)
			}
			assert.Equal(t, tt.wantCode, res.Code)
		})
	}
}
