package alloc

import (
	"testing"
)

const N = 100000

func alloc_bench() {
	for range N {
		slice := Allocate[*[100]int](8 * N)
		slice[0] = 1
		Free(slice)
	}
}

func normal_bench() {
	for range N {
		slice := make([]int, N)
		slice[0] = 1
	}
}

func BenchmarkMalloc(b *testing.B) {
	for range b.N {
		alloc_bench()
	}
}

func BenchmarkGc(b *testing.B) {
	for range b.N {
		normal_bench()
	}
}
