package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/shivam-909/gofullymanual/internal/orderbook"
	manualbook "github.com/shivam-909/gofullymanual/internal/orderbook/manual"

	_ "net/http/pprof"
)

func main() {

	ns := os.Args[1]
	N, err := strconv.Atoi(ns)
	if err != nil {
		panic(err)
	}

	ob := manualbook.New()

	start := time.Now()

	for i := 0; i < N; i++ {
		orderbook.Act(ob)
	}

	elapsed := time.Since(start)
	average := elapsed / time.Duration(N)

	fmt.Printf("Manual Allocator || %d OPS || TOTAL: %v || AVERAGE: %v\n", N, elapsed, average)
}
