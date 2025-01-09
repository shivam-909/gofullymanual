// Package alloc provides a custom memory allocator implementation
package alloc

import (
	"os"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

var (
	pageSize = os.Getpagesize()
)

const (
	wordSize = unsafe.Sizeof(uintptr(0))

	minBlockSize = 4 * wordSize

	minSplitSize = 4 * minBlockSize
)

var blockSizes = [...]int{
	64, 128, 256, 512, 1024, 2048, 4096, 8192,
}

type heap struct {
	mu sync.Mutex

	// Free list management
	freeList unsafe.Pointer
	chunks   []chunk

	// Fixed allocations management
	fixedPools map[int]unsafe.Pointer

	// Alloc tracking
	allocated int
}

// chunk represents a mapped memory region
type chunk struct {
	base unsafe.Pointer
	size int
}

// Global heap instance
var globalHeap = newHeap()

func newHeap() *heap {
	h := &heap{
		fixedPools: make(map[int]unsafe.Pointer, len(blockSizes)),
		chunks:     make([]chunk, 0, 16),
	}

	for _, size := range blockSizes {
		h.fixedPools[size] = nil
	}

	h.initializeHeap()

	return h
}

func (h *heap) initializeHeap() {
	p := h.extendHeap(pageSize)
	if p == nil {
		panic("failed to initialize heap")
	}
	h.freeList = p
}

type blockHeader struct {
	size     uintptr // Higher bits store size, lowest bit stores allocation status
	prevFree unsafe.Pointer
	nextFree unsafe.Pointer
}

func (h *heap) getHeader(p unsafe.Pointer) *blockHeader {
	return (*blockHeader)(p)
}

func (h *heap) getFooter(p unsafe.Pointer) *blockHeader {
	header := h.getHeader(p)
	size := header.size &^ 1
	return (*blockHeader)(unsafe.Add(p, int(size+wordSize)))
}

func (h *heap) isAllocated(p unsafe.Pointer) bool {
	return h.getHeader(p).size&1 != 0
}

func (h *heap) setAllocated(p unsafe.Pointer, allocated bool) {
	header := h.getHeader(p)
	if allocated {
		header.size |= 1
	} else {
		header.size &^= 1
	}
	footer := h.getFooter(p)
	footer.size = header.size
}

func (h *heap) blockSize(p unsafe.Pointer) uintptr {
	return h.getHeader(p).size &^ 1
}

// Memory mapping operations
func (h *heap) mmap(size int) unsafe.Pointer {
	size = ((size + pageSize - 1) / pageSize) * pageSize

	ptr, _, errno := syscall.Syscall6(
		syscall.SYS_MMAP,
		0,
		uintptr(size),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_PRIVATE|syscall.MAP_ANON,
		^uintptr(0),
		0,
	)

	if errno != 0 {
		return nil
	}

	return unsafe.Pointer(ptr)
}

func (h *heap) munmap(p unsafe.Pointer, size int) {
	_, _, errno := syscall.Syscall(
		syscall.SYS_MUNMAP,
		uintptr(p),
		uintptr(size),
		0,
	)
	if errno != 0 {
		panic(errno)
	}
}

func (h *heap) findFreeBlock(size int) unsafe.Pointer {
	if block := h.findInFixedPools(size); block != nil {
		return block
	}

	for p := h.freeList; p != nil; p = h.getHeader(p).nextFree {
		if h.blockSize(p) >= uintptr(size) {
			return p
		}
	}

	return h.extendHeap(size)
}

func (h *heap) findInFixedPools(size int) unsafe.Pointer {
	for _, threshold := range blockSizes {
		if size <= threshold && size >= int(float64(threshold)*0.8) {
			if block := h.fixedPools[threshold]; block != nil {
				h.removeFromFreeList(block)
				h.setAllocated(block, true)
				return block
			}
		}
	}
	return nil
}

func (h *heap) extendHeap(size int) unsafe.Pointer {
	size = ((size + pageSize - 1) / pageSize) * pageSize

	p := h.mmap(size)
	if p == nil {
		return nil
	}

	// Initialize as one large free block
	header := h.getHeader(p)
	header.size = uintptr(size - 2*int(wordSize))
	h.setAllocated(p, false)

	h.chunks = append(h.chunks, chunk{p, size})
	h.insertAtFreeListHead(p)

	return p
}

func (h *heap) removeFromFreeList(p unsafe.Pointer) {
	header := h.getHeader(p)

	if header.prevFree != nil {
		h.getHeader(header.prevFree).nextFree = header.nextFree
	} else {
		h.freeList = header.nextFree
	}

	if header.nextFree != nil {
		h.getHeader(header.nextFree).prevFree = header.prevFree
	}
}

func (h *heap) insertAtFreeListHead(p unsafe.Pointer) {
	header := h.getHeader(p)
	header.prevFree = nil
	header.nextFree = h.freeList

	if h.freeList != nil {
		h.getHeader(h.freeList).prevFree = p
	}

	h.freeList = p
}

// Allocate allocates memory for type T with the given size.
// The size represents the number of bytes, not number of items
// of type T to allocate.
func Allocate[T any](size int) *T {
	if size < int(minBlockSize) {
		size = int(minBlockSize)
	}

	globalHeap.mu.Lock()
	defer globalHeap.mu.Unlock()

	// Align size
	size = (size + int(wordSize) - 1) &^ (int(wordSize) - 1)

	block := globalHeap.findFreeBlock(size)
	if block == nil {
		return nil
	}

	globalHeap.allocateBlock(block, size)
	globalHeap.allocated++

	// Return payload area
	return (*T)(unsafe.Add(block, int(wordSize)))
}

// AllocateSlice allocates a slice of type T with the given length.
// The length represents tehe number of elements, not the number of bytes.
// Any attempt to grow the slices allocated here will cause
// undefined behaviour.
func AllocateSlice[T any](length int) []T {
	size := length * int(unsafe.Sizeof(*new(T)))
	dataPtr := Allocate[T](size)
	if dataPtr == nil {
		return nil
	}
	return unsafe.Slice(dataPtr, length)
}

// Free frees previously allocated memory
// Does nothing if the memory is already free.
func Free[T any](ptr *T) {
	if ptr == nil {
		return
	}

	globalHeap.mu.Lock()
	defer globalHeap.mu.Unlock()

	p := unsafe.Add(unsafe.Pointer(ptr), -int(wordSize))
	if !globalHeap.isAllocated(p) {
		return
	}

	globalHeap.setAllocated(p, false)
	globalHeap.coalesceBlock(p)

	globalHeap.allocated--
	if globalHeap.allocated == 0 {
		globalHeap.reset()
	}
}

// FreeSlice frees a previously allocated slice
func FreeSlice[T any](slice []T) {
	if len(slice) == 0 {
		return
	}
	Free(&slice[0])
}

// Helper methods

func (h *heap) allocateBlock(p unsafe.Pointer, size int) {
	h.removeFromFreeList(p)
	h.setAllocated(p, true)

	oldSize := h.blockSize(p)
	if oldSize >= uintptr(size)+minSplitSize {
		// Split block if enough space
		newSize := uintptr(size)
		h.getHeader(p).size = newSize | 1

		// Setup remainder as free block
		remainder := unsafe.Add(p, size+int(2*wordSize))
		h.getHeader(remainder).size = oldSize - newSize - 2*wordSize
		h.setAllocated(remainder, false)
		h.coalesceBlock(remainder)
	}
}

func (h *heap) coalesceBlock(p unsafe.Pointer) {
	// We only care for the previous and next block
	// Since the assumption is they have already been
	// coalesced.

	// Try to merge with next block
	next := unsafe.Add(p, int(h.blockSize(p)+2*wordSize))
	if h.isValidBlock(next) && !h.isAllocated(next) {
		h.removeFromFreeList(next)
		newSize := h.blockSize(p) + h.blockSize(next) + 2*wordSize
		h.getHeader(p).size = newSize
		h.getFooter(p).size = newSize
	}

	// Try to merge with previous block
	prev := h.getPreviousBlock(p)
	if prev != nil && !h.isAllocated(prev) {
		h.removeFromFreeList(prev)
		newSize := h.blockSize(prev) + h.blockSize(p) + 2*wordSize
		h.getHeader(prev).size = newSize
		h.getFooter(prev).size = newSize
		p = prev
	}

	h.insertAtFreeListHead(p)
}

func (h *heap) isValidBlock(p unsafe.Pointer) bool {
	ptr := uintptr(p)
	for _, chunk := range h.chunks {
		start := uintptr(chunk.base)
		end := start + uintptr(chunk.size)
		if ptr >= start && ptr < end {
			return true
		}
	}
	return false
}

func (h *heap) getPreviousBlock(p unsafe.Pointer) unsafe.Pointer {
	if !h.isValidBlock(unsafe.Add(p, -int(wordSize))) {
		return nil
	}

	prevFooter := (*blockHeader)(unsafe.Add(p, -int(wordSize)))
	prevSize := prevFooter.size &^ 1
	if prevSize == 0 {
		return nil
	}

	prev := unsafe.Add(p, -int(prevSize+2*wordSize))
	if !h.isValidBlock(prev) {
		return nil
	}
	return prev
}

func (h *heap) reset() {
	for _, chunk := range h.chunks {
		h.munmap(chunk.base, chunk.size)
	}
	h.chunks = h.chunks[:0]
	h.freeList = nil
	h.initializeHeap()
}

func init() {
	// Set finalizer to ensure cleanup
	runtime.SetFinalizer(globalHeap, func(h *heap) {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.reset()
	})
}
