package alloc

import (
	"fmt"
	"testing"
)

func TestSizeClass(t *testing.T) {
	input := uintptr(48)
	class := sizeToSizeClass(input)
	size := sizeClassToSize(class)
	fmt.Printf("Input: %d, Class: %d, Size: %d\n", input, class, size)
	if size < input {
		fmt.Printf("ERROR: Output size %d is less than input size %d\n", size, input)
	}
}
