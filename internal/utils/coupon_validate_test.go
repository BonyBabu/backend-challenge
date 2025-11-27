package utils_test

import (
	"backend-challenge/internal/utils"
	"fmt"
	"path"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateLineIndex(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("Failed to get current file path")
	}
	dummyCoupons := path.Join(path.Dir(filename), "/testdata/dummy_coupons")
	tests := []struct {
		name           string // description of this test case
		numberOfChunks int
		wantErr        bool
		dummyCoupons   []string
	}{
		{
			name:           "valid coupon file",
			numberOfChunks: 4,
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := utils.SplitFileToNchunks(dummyCoupons, tt.numberOfChunks)
			assert.Equal(t, len(result), tt.numberOfChunks)
			assert.Equal(t, result[0][0], int64(0))
			assert.Equal(t, result[0][1], int64(46))
			assert.Equal(t, result[1][0], int64(46))
			assert.Equal(t, result[1][1], int64(93))
			assert.Equal(t, result[2][0], int64(93))
			assert.Equal(t, result[2][1], int64(135))
			assert.Equal(t, result[3][0], int64(135))
			assert.Equal(t, result[3][1], int64(146))
		})
	}
}

func TestReadFile(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		filePath        string
		numberOfThreads int64
		expecedCoupons  []string
	}{
		{
			name:            "valid coupon file",
			filePath:        "/testdata/dummy_coupons",
			numberOfThreads: 4,
			expecedCoupons: []string{"112345678",
				"2123456789",
				"31234567891011",
				"412345678",
				"51234567891011",
				"6123456789",
				"712345678",
				"8123456789",
				"91234567891011",
				"1012345678",
				"111234567891011",
				"12123456789"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, filename, _, ok := runtime.Caller(0)
			if !ok {
				t.Fatalf("Failed to get current file path")
			}
			testCoupons := path.Join(path.Dir(filename), tt.filePath)
			coupons := map[string]any{}
			couponQueue := make(chan string, 1)
			wg, error := utils.ReadFile(testCoupons, tt.numberOfThreads, couponQueue, &atomic.Bool{})
			assert.Equal(t, error, nil)
			go func() {
				wg.Wait()          // Wait for all sender goroutines to finish
				close(couponQueue) // Close the channel
				fmt.Println("Channel closed.")
			}()
			for {
				coupon, more := <-couponQueue
				if !more {
					break
				}
				coupons[coupon] = nil
			}
			for _, coupon := range tt.expecedCoupons {
				_, exists := coupons[coupon]
				assert.Equal(t, exists, true)
			}

		})
	}
}
