package vm

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	"github.com/dominant-strategies/go-quai/params"
)

func TestTinyGoHello(t *testing.T) {
	var (
		env             = NewEVM(BlockContext{}, TxContext{}, nil, params.TestChainConfig, Config{})
		wasmInterpreter = NewWASMInterpreter(env, env.Config)
	)

	wasmBytes, err := ioutil.ReadFile("wasm_contracts/hello/hello.wasm")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading wasm file:", err)
		os.Exit(1)
	}

	// Track time taken and memory usage
	// defer trackTime(time.Now(), "wasm")

	// Initialise a new contract and set the code that is to be used by the EVM.
	// The contract is a scoped environment for this execution context only.
	contract := NewContract(&dummyContractRef{}, &dummyContractRef{}, new(big.Int), 2000)
	contract.SetCodeOptionalHash(nil, &codeAndHash{
		code: wasmBytes,
	})

	_, err2 := wasmInterpreter.Run(contract, nil, false)
	if err2 != nil {
		t.Errorf("error: %v", err)
	}
}

func TestTinyGoWRC20(t *testing.T) {
	var (
		env             = NewEVM(BlockContext{}, TxContext{}, nil, params.TestChainConfig, Config{})
		wasmInterpreter = NewWASMInterpreter(env, env.Config)
	)

	wasmBytes, err := ioutil.ReadFile("wasm_contracts/wrc-20/wrc-20.wasm")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading wasm file:", err)
		os.Exit(1)
	}

	// Track time taken and memory usage
	// defer trackTime(time.Now(), "wasm")

	// Initialise a new contract and set the code that is to be used by the EVM.
	// The contract is a scoped environment for this execution context only.
	contract := NewContract(&dummyContractRef{}, &dummyContractRef{}, new(big.Int), 2000)
	contract.SetCodeOptionalHash(nil, &codeAndHash{
		code: wasmBytes,
	})

	_, err2 := wasmInterpreter.Run(contract, nil, false)
	if err2 != nil {
		t.Errorf("error: %v", err)
	}
}
