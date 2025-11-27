package utils

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"sync"
	"sync/atomic"
)

// type Set = *ConcurrentMap[string, any]

func ReadFile(filePath string, numberOfThreads int64, couponQueue chan<- string, stopReading *atomic.Bool) (*sync.WaitGroup, error) {
	chunks, err := SplitFileToNchunks(filePath, int(numberOfThreads))
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	for _, chunk := range chunks {
		wg.Add(1)
		go ReadFileChunk(filePath, chunk[0], chunk[1], couponQueue, &wg, stopReading)
	}
	return &wg, nil
}

func ReadFileChunk(filePath string, from int64, to int64, couponQueue chan<- string, wg *sync.WaitGroup, stopReading *atomic.Bool) error {
	defer wg.Done()
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	buf := make([]byte, to-from)
	_, err = file.ReadAt(buf, from)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(bytes.NewReader(buf))
	for scanner.Scan() {
		if stopReading.Load() {
			break
		}
		line := scanner.Text()
		couponQueue <- line
	}
	return nil
}

func SplitFileToNchunks(filePath string, n int) ([][]int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := fileInfo.Size()
	medianChunkSize := fileSize / int64(n)
	var chunks [][]int64
	var start int64 = 0
	for i := range n {
		var end int64
		if i == n-1 {
			end = fileSize
		} else {
			end = start + medianChunkSize
			// Move end to the next newline character
			buf := make([]byte, 1)
			for {
				_, err := file.ReadAt(buf, end)
				if err != nil {
					end = fileSize
					break
				}
				if buf[0] == '\n' {
					end++
					break
				}
				end++
			}
		}
		chunks = append(chunks, []int64{start, end})
		start = end
	}

	return chunks, nil
}

func ScanForCoupon(numberOfThreads int64, couponQueue <-chan string, expectedCoupon string, file string, stopProducers *atomic.Bool) (*atomic.Bool, *sync.WaitGroup) {
	var wgRecivers sync.WaitGroup
	var flag atomic.Bool
	for i := 0; i < int(numberOfThreads); i++ {
		wgRecivers.Add(1)
		go func() {
			defer wgRecivers.Done()
			for coupon := range couponQueue {
				if coupon == expectedCoupon {
					flag.Store(true)
					log.Println("Found", expectedCoupon, "in", file)
					break
				}
			}
		}()
	}
	return &flag, &wgRecivers
}
