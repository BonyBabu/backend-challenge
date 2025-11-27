package db

import (
	"backend-challenge/internal/generated/openapi"
	"backend-challenge/internal/utils"
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
)

type couponDaoImpl struct {
	files     []string
	couponMin int
}

func NewCouponDao(files []string, couponMin int) CouponDao {
	return &couponDaoImpl{files: files, couponMin: couponMin}
}

type SearchResultImpl struct {
	resultSet        []*atomic.Bool
	searchProcessors []*sync.WaitGroup
	stopAllchecks    chan bool
	couponMin        int
	retrieved        bool
}

// Validate implements SearchResult.
func (s *SearchResultImpl) Validate() (bool, error) {
	if s.retrieved {
		return false, errors.New("already retrived")
	}
	s.retrieved = true
	counter := 0
	var stopWaitingForResult atomic.Bool
	stopWaitingForResult.Store(false)
	go func() {
		for _, fileProcessor := range s.searchProcessors {
			fileProcessor.Wait()
			stopWaitingForResult.Store(true)
		}
	}()
	for {
		for i, result := range s.resultSet {
			if result.Load() {
				counter = counter + 1
				s.resultSet[i].Store(false)
			}
		}
		if counter == s.couponMin || stopWaitingForResult.Load() {
			s.stopAllchecks <- true
			break
		}
	}
	return counter >= s.couponMin, nil
}

func searchForCoupon(filePath string, numberOfThreads int64, coupon string, stopAllchecks <-chan bool) (*atomic.Bool, *sync.WaitGroup, error) {
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

// SearchForCouponInGivenFiles implements CouponDao.
func (c *couponDaoImpl) SearchForCouponInGivenFiles(ctx context.Context, orderReq openapi.OrderReq) (SearchResult, error) {
	stopAllchecks := make(chan bool, 1)
	// kill all threads if context is canccelled
	go func() {
		result := ctx.Done()
		if result != nil {
			<-result
			stopAllchecks <- true
		}
	}()
	var results []*atomic.Bool
	var fileProcessors []*sync.WaitGroup
	for _, file := range c.files {
		flag, wg, err := searchForCoupon(file, 10, orderReq.CouponCode, stopAllchecks)
		if err != nil {
			return nil, err
		}
		results = append(results, flag)
		fileProcessors = append(fileProcessors, wg)
	}
	return &SearchResultImpl{results, fileProcessors, stopAllchecks, c.couponMin, false}, nil
}

var _ CouponDao = &couponDaoImpl{}
