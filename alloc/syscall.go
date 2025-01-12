package alloc

import (
	"syscall"
	"unsafe"
)

func mmap(size int) unsafe.Pointer {
	if globalArena != nil && (uintptr(globalArena)+uintptr(globalOffset)) < arenaSize {
		ret := unsafe.Add(globalArena, globalOffset)
		globalOffset += size
		return ret
	}

	prot := uintptr(syscall.PROT_READ | syscall.PROT_WRITE)
	flags := uintptr(syscall.MAP_PRIVATE | syscall.MAP_ANON)
	fd := uintptr(^uint64(0)) // -1
	offset := uintptr(0)

	addr, _, errno := syscall.Syscall6(
		syscall.SYS_MMAP,
		0,
		uintptr(arenaSize),
		prot,
		flags,
		fd,
		offset,
	)

	globalArena = unsafe.Pointer(addr)
	globalOffset = size

	if int64(addr) == -1 {
		panic(errno)
	}

	return globalArena
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
