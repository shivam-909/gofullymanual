package alloc

import (
	"fmt"
	"runtime"
	"testing"
)

func TestLeak(t *testing.T) {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	fmt.Printf("BEFORE ALLOC: Alloc = %v\n", before.HeapSys)

	const numAllocs = 1
	ptrs := make([]*int, 0, numAllocs)
	for i := 0; i < numAllocs; i++ {
		p := Allocate[int](1000) // allocate 1000 bytes
		if p == nil {
			t.Fatalf("failed to allocate on iteration %d", i)
		}
		ptrs = append(ptrs, p)
	}

	runtime.ReadMemStats(&after)
	fmt.Printf("BEFORE FREE:  Alloc = %v\n", after.HeapSys)

	for _, p := range ptrs {
		Free(p)
	}

	runtime.GC()

	runtime.ReadMemStats(&after)
	fmt.Printf("AFTER FREE:  Alloc = %v\n", after.HeapSys)

	if after.Alloc > before.Alloc*2 {
		t.Errorf("Possible leak: memory usage after = %d, before = %d", after.Alloc, before.Alloc)
	}
}
