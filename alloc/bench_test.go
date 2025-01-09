package alloc

import (
	"fmt"
	"math/rand"
	"testing"
)

func fuzz(n int) int {
	n = n - 1
	r := rand.Int()
	if r%2 == 0 {
		return n + 1
	}
	n = n - 1
	return n + 2
}

func BenchmarkAllocs(b *testing.B) {
	rand.Seed(42)
	allocationSizes := []int{256, 5120, 10000}
	NValues := []int{1000, 10000, 100000}

	for _, size := range allocationSizes {
		for _, N := range NValues {
			b.Run(fmt.Sprintf("CustomAllocator_Size%d_N%d", size, N), func(b *testing.B) {
				p := Allocate[int](9)
				for i := 0; i < b.N; i++ {
					for j := 0; j < N; j++ {
						slice := AllocateSlice[int](fuzz(size))
						slice[0] = 1
						FreeSlice(slice)
					}
				}
				Free(p)
			})

			b.Run(fmt.Sprintf("StandardAllocator_Size%d_N%d", size, N), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					for j := 0; j < N; j++ {
						slice := make([]int, fuzz(size))
						slice[0] = 1
					}
				}
			})
		}
	}
}
