package alloc

import (
	"testing"
)

const N = 10000

func BenchmarkLargeAllocs(b *testing.B) {
	b.Run("CustomAllocator", func(b *testing.B) {
		p := Allocate[int](8)

		for i := 0; i < b.N; i++ {
			for j := 0; j < N; j++ {
				slice := AllocateSlice[int](10000)
				slice[0] = 1
				FreeSlice(slice)
			}
		}
		Free(p)
	})

	b.Run("StandardAllocator", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < N; j++ {
				slice := make([]int, 10000)
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
				slice := AllocateSlice[int](5120)
				slice[0] = 1
				FreeSlice(slice)
			}
		}
		Free(p)
	})

	b.Run("StandardAllocator", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < N; j++ {
				slice := make([]int, 5120)
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
				slice := AllocateSlice[int](256)
				slice[0] = 1
				FreeSlice(slice)
			}
		}
		Free(p)
	})

	b.Run("StandardAllocator", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < N; j++ {
				slice := make([]int, 256)
				slice[0] = 1
			}
		}
	})
}
