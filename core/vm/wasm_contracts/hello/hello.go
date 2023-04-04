package main

import "fmt"

func main() {}

//export run
//go:linkname run
func run() {
	fmt.Println("Hello, World!")
}