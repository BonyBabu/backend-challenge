package services

import (
	"backend-challenge/internal/db"
	openapi "backend-challenge/internal/generated/openapi"
	"backend-challenge/internal/utils"
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

// OrderAPIService is a service that implements the logic for the OrderAPIServicer
// This service should implement the business logic for every endpoint for the OrderAPI API.
// Include any external packages or services that will be required by this service.
type OrderAPIService struct {
	orderDao   db.OrderDao
	productDao db.ProductDao
	couponDao  db.CouponDao
}

func SearchForCoupon(filePath string, numberOfThreads int64, coupon string, stopAllchecks <-chan bool) (*atomic.Bool, *sync.WaitGroup, error) {
	var stopProducers atomic.Bool
	stopProducers.Store(false)
	go func() {
		val, ok := <-stopAllchecks
		if ok {
			stopProducers.Store(val)
		}
	}()
	couponQueue := make(chan string, numberOfThreads*100)
	wgProducers, err := utils.ReadFile(filePath, numberOfThreads, couponQueue, &stopProducers)
	if err != nil {
		return nil, nil, err
	}
	go func() {
		wgProducers.Wait() // Wait for all sender goroutines to finish
		close(couponQueue) // Close the channel
		log.Printf("Stop processing file %s", filePath)
	}()
	atomicBool, wgRecivers := utils.ScanForCoupon(numberOfThreads, couponQueue, coupon, filePath, &stopProducers)
	return atomicBool, wgRecivers, nil
}

// NewOrderAPIService creates a default api service
func NewOrderAPIService(orderDao db.OrderDao, productDao db.ProductDao, files []string, couponMin int) *OrderAPIService {
	return &OrderAPIService{
		orderDao:   orderDao,
		productDao: productDao,
		couponDao:  db.NewCouponDao(files, couponMin),
	}
}

// NewOrderAPIServiceWithCouponDao creates a default api service by injecting couponDao directly.
// This constructor is useful for unit tests where a couponDao mock can be provided.
func NewOrderAPIServiceWithCouponDao(orderDao db.OrderDao, productDao db.ProductDao, couponDao db.CouponDao) *OrderAPIService {
	return &OrderAPIService{
		orderDao:   orderDao,
		productDao: productDao,
		couponDao:  couponDao,
	}
}

// PlaceOrder - Place an order
func (s *OrderAPIService) PlaceOrder(ctx context.Context, orderReq openapi.OrderReq) (res openapi.ImplResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("thread is recovered")
			err = errors.ErrUnsupported
			res = openapi.ImplResponse{}
		}
	}()

	if len(orderReq.CouponCode) < 8 || len(orderReq.CouponCode) > 10 {
		return openapi.Response(http.StatusUnprocessableEntity, "invalid coupon code"), nil
	}

	searchResult, err := s.couponDao.SearchForCouponInGivenFiles(orderReq)
	if err != nil {
		return openapi.ImplResponse{}, nil
	}

	id := uuid.New().String()
	items := make([]openapi.OrderItemsInner, 0, len(orderReq.Items))
	products := make([]openapi.Product, 0, len(orderReq.Items))

	for _, item := range orderReq.Items {
		if item.Quantity <= 0 {
			return openapi.Response(http.StatusBadRequest, "quantity must be greater than zero"), nil
		}
		product, err := s.productDao.GetProduct(ctx, db.ID(item.ProductId))
		if err != nil {
			return openapi.Response(http.StatusBadRequest, "invalid product specified"), nil
		}
		// Convert OrderReqItemsInner to OrderItemsInner using type conversion
		items = append(items, openapi.OrderItemsInner(item))
		openapiProduct := openapi.Product{
			Id:       product.Id,
			Name:     product.Name,
			Price:    product.Price,
			Category: product.Category,
		}
		products = append(products, openapiProduct)
	}

	if result, err := searchResult.Validate(); err != nil {
		return openapi.ImplResponse{}, err
	} else if !result {
		return openapi.Response(http.StatusUnprocessableEntity, "invalid coupon code"), nil
	}

	if err = s.orderDao.CreateOrder(ctx, db.Order{
		ID: id,
		Items: func() []db.Item {
			var dbItems []db.Item
			for _, item := range orderReq.Items {
				dbItems = append(dbItems, db.Item{
					ProductID: db.ID(item.ProductId),
					Quantity:  item.Quantity,
				})
			}
			return dbItems
		}(),
	}); err != nil {
		return openapi.ImplResponse{}, err
	}

	return openapi.Response(http.StatusOK, openapi.Order{
		Id:       id,
		Items:    items,
		Products: products,
	}), nil
}
