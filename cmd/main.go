package main

import (
	"github.com/shivam-909/gofullymanual/alloc"

	_ "net/http/pprof"
)

const N = 100000000

func main() {
	slice := alloc.Allocate[*[N]int](8 * N)
	for i := range N {
		slice[i] = i
	}

	alloc.Free(slice)
}
