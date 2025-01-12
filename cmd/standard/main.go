package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/shivam-909/gofullymanual/internal/orderbook"
	standardbook "github.com/shivam-909/gofullymanual/internal/orderbook/standard"
)

func main() {

	ns := os.Args[1]
	N, err := strconv.Atoi(ns)
	if err != nil {
		panic(err)
	}

	ob := standardbook.New()

	start := time.Now()

	for i := 0; i < N; i++ {
		orderbook.Act(ob)
	}

	elapsed := time.Since(start)
	average := elapsed / time.Duration(N)

	fmt.Printf("Standard Allocator || %d OPS || TOTAL: %v || AVERAGE: %v\n", N, elapsed, average)
}
