package main

// import "unsafe"

// // main is required for TinyGo to compile to Wasm.
// func main() {}

// //export run
// //go:linkname run
// func run() {
// 	var address [8]uint32
// 	rawAddress := uintptr(unsafe.Pointer(&address[0]))
// 	int32Address := (*int32)(unsafe.Pointer(rawAddress))
// 	getAddress(int32Address)
// }

// //export getAddress
// //go:linkname getAddress
// func getAddress(resultOffset int32)

/*
#include <stdint.h>

extern void getAddress(uint32_t *offset);
extern void storageStore(uint32_t *keyOffset, uint32_t *valueOffset);
*/
import (
	"fmt"
)

func main() {}

//export run
//go:linkname run
func run() {
	address := getAddress()

	fmt.Println(address)
}

//export getAddress
//go:linkname getAddress
func getAddress() string
