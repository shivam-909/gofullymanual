package alloc

import (
	"math/rand"
	"testing"
)

const N = 1000

func fuzz(n int) int {
	n = n - 1

	r := rand.Int()
	if r%2 == 0 {
		return n + 1
	}

	n = n - 1

	return n + 2
}

func BenchmarkLargeAllocs(b *testing.B) {
	b.Run("CustomAllocator", func(b *testing.B) {
		p := Allocate[int](8)

		for i := 0; i < b.N; i++ {
			for j := 0; j < N; j++ {
				slice := AllocateSlice[int](fuzz(10000))
				slice[0] = 1
				FreeSlice(slice)
			}
		}
		Free(p)
	})

	b.Run("StandardAllocator", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < N; j++ {
				slice := make([]int, fuzz(10000))
				slice[0] = 1
			}
		}
	})
}

func BenchmarkSmallAllocs(b *testing.B) {
	b.Run("CustomAllocator", func(b *testing.B) {
		p := Allocate[int](8)
		for i := 0; i < b.N; i++ {
			for j := 0; j < N; j++ {
				slice := AllocateSlice[int](fuzz(5120))
				slice[0] = 1
				FreeSlice(slice)
			}
		}
		Free(p)
	})

	b.Run("StandardAllocator", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < N; j++ {
				slice := make([]int, fuzz(5120))
				slice[0] = 1
			}
		}
	})
}

func BenchmarkTinyAllocs(b *testing.B) {
	b.Run("CustomAllocator", func(b *testing.B) {
		p := Allocate[int](8)
		for i := 0; i < b.N; i++ {
			for j := 0; j < N; j++ {
				slice := AllocateSlice[int](fuzz(256))
				slice[0] = 1
				FreeSlice(slice)
			}
		}
		Free(p)
	})

	b.Run("StandardAllocator", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < N; j++ {
				slice := make([]int, fuzz(256))
				slice[0] = 1
			}
		}
	})
}
