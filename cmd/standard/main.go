package main

import (
	"fmt"
	"time"

	"github.com/shivam-909/gofullymanual/internal/orderbook"
	standardbook "github.com/shivam-909/gofullymanual/internal/orderbook/standard"
)

func main() {
	ob := standardbook.New()
	N := 2500000

	start := time.Now()

	for i := 0; i < N; i++ {
		orderbook.Act(ob)
	}

	elapsed := time.Since(start)
	average := elapsed / time.Duration(N)

	fmt.Printf("Total time for %d operations: %v\n", N, elapsed)
	fmt.Printf("Average time per operation: %v\n", average)
}
