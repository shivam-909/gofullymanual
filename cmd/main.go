package main

import (
	"fmt"

	"github.com/shivam-909/gofullymanual/alloc"

	_ "net/http/pprof"
)

const N = 10

func main() {
	slice := alloc.AllocateSlice[int](N)
	for i := range N {
		slice[i] = i
	}

	for _, v := range slice {
		fmt.Println(v)
	}

	alloc.FreeSlice(slice)
}
