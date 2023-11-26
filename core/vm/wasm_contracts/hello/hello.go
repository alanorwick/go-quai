package main

// main is required for TinyGo to compile to Wasm.
func main() {}

//export main
func run() {
	hello()
}

// hello is a simple function with no host calls.
func hello() {
	// Perform your logic here. Remember that you cannot use
	// functions like fmt.Println to output to the console,
	// as they are host functions.
	// Instead, you can implement logic that can be observed
	// via WebAssembly's memory or exported functions.
}
