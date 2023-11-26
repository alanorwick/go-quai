package main

//export getCallDataSize
//go:linkname getCallDataSize
func getCallDataSize() int32

// main is required for TinyGo to compile to Wasm.
func main() {}

// Exported function that routes the call based on the selector.
//
//export main
func run() {
	if getCallDataSize() < 4 {
		revert(0, 0)
	}

	selector := getSelector()
	switch selector {
	case 0x9993021a:
		doBalance()
	case 0x5d359fbd:
		doTransfer()
	default:
		revert(0, 0)
	}
}

func doBalance() {
	if getCallDataSize() != 24 {
		revert(0, 0)
	}

	// address := getCallDataAddress(4)
	// balance := getBalance(address)
	// finish(balance.Bytes(), 32)
}

func doTransfer() {
	if getCallDataSize() != 32 {
		revert(0, 0)
	}

	sender := getCaller()
	recipient := getCallDataAddress(4)
	// value := getCallDataValue(24)

	senderBalance := getBalance(sender)
	recipientBalance := getBalance(recipient)

	// if senderBalance.Cmp(value) < 0 {
	// 	revert(0, 0)
	// }

	// senderBalance.Sub(senderBalance, value)
	// recipientBalance.Add(recipientBalance, value)

	setBalance(sender, senderBalance)
	setBalance(recipient, recipientBalance)
}

// Helper functions to interact with storage and call data.
func getSelector() uint32 {
	// Extract the first 4 bytes of call data as the function selector.
	// ...
	return 0
}

func getCallDataAddress(offset int32) []byte {
	// Extract an address from call data at the given offset.
	// ...
	return []byte{}
}

func getCallDataValue(offset int32) int32 {
	// Extract a value from call data at the given offset.
	// ...
	return 0
}

func getBalance(address []byte) int32 {
	// Use storageLoad to get the balance for the address.
	// ...
	return 0
}

func setBalance(address []byte, balance int32) {
	// Use storageStore to set the balance for the address.
	// ...
}

func getCaller() []byte {
	// Get the address of the caller.
	// ...
	return []byte{}
}

func revert(offset int32, length int32) {
	// ...
}

func finish(offset int32, length int32) {
	// ...
}
