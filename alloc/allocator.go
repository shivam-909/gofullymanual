package alloc

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	chunkSize = os.Getpagesize()          // typically 4096
	dWordSize = unsafe.Sizeof(uintptr(0)) // size of one word (8 bytes on 64-bit)
)

type chunk struct {
	base unsafe.Pointer
	size int
}

// Global free list management
var (
	heaplist  unsafe.Pointer = nil // head of free list
	allocated int
	chunks    []chunk = make([]chunk, 0) // we kindly ask go for just this one heap allocation
)

func mmap(length int) unsafe.Pointer {
	prot := uintptr(syscall.PROT_READ | syscall.PROT_WRITE)
	flags := uintptr(syscall.MAP_PRIVATE | syscall.MAP_ANON)
	fd := uintptr(^uint64(0)) // -1
	offset := uintptr(0)

	addr, _, errno := syscall.Syscall6(
		syscall.SYS_MMAP,
		0,
		uintptr(length),
		prot,
		flags,
		fd,
		offset,
	)

	if int64(addr) == -1 {
		panic(errno)
	}

	return unsafe.Pointer(addr)
}

func boolBit(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}

func bitBool(i uintptr) bool {
	return i != 0
}

// Block layout & metadata
//
// [HEADER=8 bytes][PAYLOAD...][FOOTER=8 bytes]
//
// For a free block, the payload's first two words store
// prevFree and nextFree pointers

type blockMetadata struct {
	// info has lower 3 bits for flags, upper bits for size
	info uintptr
}

func setMd(h unsafe.Pointer, size uintptr, allocated bool) {
	md := (*blockMetadata)(h)
	md.info = size | boolBit(allocated)
}

func blockSize(h unsafe.Pointer) uintptr {
	md := (*blockMetadata)(h)
	return md.info & ^(uintptr(0x7))
}

func blockAllocated(h unsafe.Pointer) bool {
	md := (*blockMetadata)(h)
	return bitBool(md.info & 0x1)
}

func byteOffset(h unsafe.Pointer, bytes int) unsafe.Pointer {
	return unsafe.Add(h, bytes)
}

// Returns pointer to the data area (after header)
func blockPayload(h unsafe.Pointer) unsafe.Pointer {
	return byteOffset(h, int(dWordSize))
}

// Returns pointer to the footer’s metadata
func blockFooter(h unsafe.Pointer) unsafe.Pointer {
	return byteOffset(h, int(dWordSize)+int(blockSize(h)))
}

func setBlockStatus(h unsafe.Pointer, allocated bool) {
	s := blockSize(h)
	f := blockFooter(h)
	setMd(h, s, allocated)
	setMd(f, s, allocated)
}

func blockPrevFree(h unsafe.Pointer) unsafe.Pointer {
	payload := blockPayload(h)
	return *(*unsafe.Pointer)(payload)
}

func blockNextFree(h unsafe.Pointer) unsafe.Pointer {
	payload := blockPayload(h)
	return *(*unsafe.Pointer)(byteOffset(payload, int(dWordSize)))
}

func setBlockPrevFree(h unsafe.Pointer, newPrev unsafe.Pointer) {
	payload := blockPayload(h)
	*(*unsafe.Pointer)(payload) = newPrev
}

func setBlockNextFree(h unsafe.Pointer, newNext unsafe.Pointer) {
	payload := blockPayload(h)
	*(*unsafe.Pointer)(byteOffset(payload, int(dWordSize))) = newNext
}

func inHeapBounds(h unsafe.Pointer) bool {
	for _, chunk := range chunks {
		heapBase := chunk.base
		start := uintptr(heapBase)
		end := start + uintptr(chunk.size)
		ptr := uintptr(h)
		if ptr >= start && ptr < end {
			return true
		}
	}
	return false
}

func nextAdjBlock(h unsafe.Pointer) unsafe.Pointer {
	n := byteOffset(h, int(blockSize(h)+2*dWordSize))
	if !inHeapBounds(n) {
		return nil
	}
	return n
}

func prevAdjBlock(h unsafe.Pointer) unsafe.Pointer {
	// The “footer” of the previous block is right before this header
	footerPtr := byteOffset(h, -int(dWordSize))
	if !inHeapBounds(footerPtr) {
		return nil
	}

	prevSize := (*blockMetadata)(footerPtr).info & ^(uintptr(0x7))
	if prevSize == 0 {
		return nil
	}

	// The start of that previous block’s header:
	prev := byteOffset(footerPtr, -int(prevSize+dWordSize))
	if !inHeapBounds(prev) {
		return nil
	}
	return prev
}

func removeFromFreeList(h unsafe.Pointer) {
	pf := blockPrevFree(h)
	nf := blockNextFree(h)

	if pf != nil {
		setBlockNextFree(pf, nf)
	} else {
		// h was the head
		heaplist = nf
	}

	if nf != nil {
		setBlockPrevFree(nf, pf)
	}
}

func insertAtFreeListHead(h unsafe.Pointer) {
	setBlockPrevFree(h, nil)
	setBlockNextFree(h, heaplist)
	if heaplist != nil {
		setBlockPrevFree(heaplist, h)
	}
	heaplist = h
}

// Create a free block of `size` at pointer h, mark header/footer, add to free list
func createAtFreeListHead(h unsafe.Pointer, size uintptr) {
	setMd(h, size, false)
	setMd(blockFooter(h), size, false)
	insertAtFreeListHead(h)
}

func alignedSize(size int) int {
	const alignment = 8
	return (size + alignment - 1) & ^(alignment - 1)
}

func coalesceBlock(h unsafe.Pointer) {
	n := nextAdjBlock(h)
	p := prevAdjBlock(h)

	// If next block is free, remove it and combine
	if n != nil && !blockAllocated(n) {
		removeFromFreeList(n)
		newSize := blockSize(h) + blockSize(n) + 2*dWordSize
		setMd(h, newSize, false)
		setMd(blockFooter(h), newSize, false)
	}

	// If previous block is free, remove it and combine
	if p != nil && !blockAllocated(p) {
		removeFromFreeList(p)
		newSize := blockSize(p) + blockSize(h) + 2*dWordSize
		setMd(p, newSize, false)
		setMd(blockFooter(p), newSize, false)
		h = p
	}

	insertAtFreeListHead(h)
}

func initFreeList() {

	p := extendHeap(chunkSize)
	if p == nil {
		panic("failed to init heap")
	}
	heaplist = p
}

// Allocates a new chunk via mmap, creates one big free block, returns pointer
func extendHeap(sizeNeeded int) unsafe.Pointer {
	if sizeNeeded < chunkSize {
		sizeNeeded = chunkSize
	} else {
		remainder := sizeNeeded % chunkSize
		if remainder != 0 {
			sizeNeeded += (chunkSize - remainder)
		}
	}

	p := mmap(sizeNeeded)
	if p == nil {
		return nil
	}

	createAtFreeListHead(p, uintptr(sizeNeeded)-2*dWordSize)

	chunks = append(chunks, chunk{p, sizeNeeded})

	return p
}

func findFreeBlock(size int) unsafe.Pointer {
	for c := heaplist; c != nil; c = blockNextFree(c) {
		if blockSize(c) >= uintptr(size) {
			return c
		}
	}

	// Else, grow
	return extendHeap(size)
}

func allocateBlock(h unsafe.Pointer, size int) {
	if blockAllocated(h) {
		return
	}
	removeFromFreeList(h)
	setBlockStatus(h, true)

	oldSize := blockSize(h)
	minSpare := uintptr(4 * dWordSize) // enough room for header+footer + pointers
	if oldSize >= uintptr(size)+minSpare {
		// Split
		setMd(h, uintptr(size), true)
		setMd(blockFooter(h), uintptr(size), true)

		// The leftover chunk starts after this block’s header+footer+payload
		freeBlockPtr := byteOffset(h, int(uintptr(size)+2*dWordSize))
		spare := oldSize - uintptr(size) - 2*dWordSize

		setMd(freeBlockPtr, spare, false)
		setMd(blockFooter(freeBlockPtr), spare, false)

		coalesceBlock(freeBlockPtr)
	}
}

// Allocate returns a pointer to at least `size` bytes of memory
func Allocate[T *Q, Q any](size int) T {
	asize := alignedSize(size)
	if asize < int(4*dWordSize) {
		asize = int(4 * dWordSize)
	}

	h := findFreeBlock(asize)
	if h == nil {
		return nil
	}

	allocateBlock(h, asize)

	allocated += 1

	return (T)(blockPayload(h))
}

func munmap(p unsafe.Pointer, size int) {
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

// Free a previously allocated pointer
func Free[T *Q, Q any](ptr T) {
	p := unsafe.Pointer(ptr)
	if p == nil {
		return
	}

	blockHeader := byteOffset(p, -int(dWordSize))
	if !blockAllocated(blockHeader) {
		// Already free
		return
	}

	setBlockStatus(blockHeader, false)
	coalesceBlock(blockHeader)

	allocated -= 1

	if allocated < 1 {
		for _, chunk := range chunks {
			munmap(chunk.base, chunk.size)
		}

		chunks = chunks[:0]
		heaplist = nil
		initFreeList()
	}

}

func Sizeof[T *Q, Q any](x T) int {
	return int(unsafe.Sizeof(x))
}

func init() {
	initFreeList()
}
