package main

import "fmt"

// main is required for TinyGo to compile to Wasm.
func main() {}

//export run
//go:linkname run
func run() {
	fmt.Println("useGas(1000)")
	useGas(1000)
}

//export useGas
//go:linkname useGas
func useGas(amount int64)
