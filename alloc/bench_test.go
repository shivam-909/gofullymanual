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
	allocationSizes := []int{256, 5120}
	NValues := []int{1000, 10000, 100000}

	for _, size := range allocationSizes {
		l := fuzz(size)
		for _, N := range NValues {
			b.Run(fmt.Sprintf("MyAlloc_%d-Bytes_%d-Times", size, N), func(b *testing.B) {
				p := Allocate[int](1)
				for i := 0; i < b.N; i++ {
					for j := 0; j < N; j++ {
						slice := AllocateSlice[int](l)
						slice[0] = 1
						FreeSlice(slice)
					}
				}
				Free(p, 1)
			})

			b.Run(fmt.Sprintf("StandardAlloc_%d-Bytes_%d-Times", size, N), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					for j := 0; j < N; j++ {
						slice := make([]int, fuzz(l))
						slice[0] = 1
					}
				}
			})
		}
	}
}
