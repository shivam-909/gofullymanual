package alloc

import (
	"log"
	"math/bits"
	"unsafe"
)

const (
	slabSize  = 256 * 1024
	chunkSize = 32 * 1024 * 1024
	arenaSize = 8 * chunkSize

	mediumAllocationThreshold = 64 * 1024
)

// We accept some normal heap allocations
var (
	smallSlabs  = make(map[uintptr][]*smallSlab)
	mediumSlabs = make([]*mediumSlab, 0)

	globalArena  = unsafe.Pointer(nil)
	globalOffset = 0
)

type slabNode struct {
	next *slabNode
}

type smallSlab struct {
	// Base address of this slab
	mem unsafe.Pointer

	// Slabs are allocated "bump style"
	// until we reach the end of the slab
	bumpOffset uintptr

	// A free list is maintained
	// but isn't used for allocation
	// until the bumpOffset is at it's max
	freeList *slabNode

	// Size of the objects allocated
	// in this slab
	blockSize uintptr
}

type freeBlock struct {
	size uintptr
	next *freeBlock
}

type mediumSlab struct {
	// Base address of this slab
	mem unsafe.Pointer

	// Total capacity of the slab
	slabSize uintptr

	// A free list is maintained for all
	// medium sized object allocations
	freeList *freeBlock
}

func sizeToSizeClass(size uintptr) uint8 {
	size = size - 1

	leadingZeros := uint8(bits.LeadingZeros64(uint64(size | 32)))

	e := uint8(61) - leadingZeros

	b := uint8(0)
	if e != 0 {
		b = 1
	}
	shift := 4 + e - b
	m := uint8((size >> shift) & 3)

	return (e << 2) + m
}

func sizeClassToSize(sizeclass uint8) uintptr {
	if sizeclass == 0 {
		return 0
	}

	e := sizeclass >> 2
	m := sizeclass & 3

	b := uint8(0)
	if e != 0 {
		b = 1
	}

	baseSize := uintptr(16 + m*4)
	shift := e - b

	return baseSize << shift
}

// Gets a slab with free space
// or creates a new slab
func getSlab(size uintptr) *smallSlab {
	slabs := smallSlabs[size]

	// See if we have a slab for this size
	for _, slab := range slabs {
		if slab.freeList != nil || slab.bumpOffset+size <= slabSize {
			return slab
		}
	}

	ptr := mmap(slabSize)
	if ptr == nil {
		return nil
	}

	slab := &smallSlab{
		mem:        ptr,
		bumpOffset: 0,
		freeList:   nil,
		blockSize:  size,
	}

	slabs = append(slabs, slab)
	smallSlabs[size] = slabs
	return slab
}

func allocateSmallObject(size uintptr) unsafe.Pointer {
	c := sizeToSizeClass(size)
	sizeclass := sizeClassToSize(c)
	if sizeclass == 0 {
		panic("allocateSmallObject called with a medium sized allocation")
	}

	slab := getSlab(sizeclass)

	if slab.freeList != nil {
		nextFree := slab.freeList
		slab.freeList = nextFree.next
		return unsafe.Pointer(nextFree)
	}

	if slab.bumpOffset+sizeclass <= slabSize {
		addr := unsafe.Add(slab.mem, slab.bumpOffset)
		slab.bumpOffset += sizeclass
		return addr
	}

	log.Println("size:", size)
	log.Println("sizeClass:", sizeclass)
	log.Println("slab.bumpOffset+sizeclass:", slab.bumpOffset+sizeclass)
	log.Println("slabs:", smallSlabs)
	log.Println("slabSize:", slabSize)
	log.Printf("slab: %#v\n", slab)

	panic("unable to allocate small object")
}

func findSlabForObj(slabs []*smallSlab, ptr unsafe.Pointer) *smallSlab {
	p := uintptr(ptr)
	for _, s := range slabs {
		start := uintptr(s.mem)
		end := start + slabSize
		if p >= start && p < end {
			return s
		}
	}

	return nil
}

func freeSmallObject(ptr unsafe.Pointer, size uintptr) {
	c := sizeToSizeClass(size)
	sizeclass := sizeClassToSize(c)
	if sizeclass == 0 {
		panic("freeSmallObject called with a medium sized allocation")
	}

	slabs := smallSlabs[sizeclass]

	allocatedSlab := findSlabForObj(slabs, ptr)
	if allocatedSlab == nil {
		log.Println(sizeclass, smallSlabs)
		log.Fatalf("pointer %p not accounted for by allocator", ptr)
	}

	node := (*slabNode)(ptr)
	node.next = allocatedSlab.freeList
	allocatedSlab.freeList = node
}

// Allocates enough memory for count * T items and returns a pointer
// to the newly allocated memory
func Allocate[T any](count int) *T {
	sz := uintptr(count) * unsafe.Sizeof(*new(T))

	if sz <= mediumAllocationThreshold {
		return (*T)(allocateSmallObject(sz))
	}

	panic("too large to allocate")
}

// Free a previously allocated pointer
func Free[T any](ptr *T, count int) {
	if ptr == nil {
		return
	}

	sz := uintptr(count) * unsafe.Sizeof(*new(T))

	if sz <= mediumAllocationThreshold {
		freeSmallObject(unsafe.Pointer(ptr), sz)
		return
	}

	panic("not tracked by allocator (too large)")
}

func AllocateSlice[T any](elements int) []T {
	sz := unsafe.Sizeof(*new(T)) * uintptr(elements)

	if sz <= mediumAllocationThreshold {
		ptr := Allocate[T](int(sz))
		return ([]T)(unsafe.Slice(ptr, sz))
	}

	panic("too large to allocate")
}

func FreeSlice[T any](slice []T) {
	ptr := &slice[0]
	Free(ptr, len(slice))
}

func Sizeof[T any](x T) int {
	return int(unsafe.Sizeof(x))
}
