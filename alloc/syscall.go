package alloc

import (
	"syscall"
	"unsafe"
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
