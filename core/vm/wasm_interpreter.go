package vm

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"

	"github.com/second-state/WasmEdge-go/wasmedge"
	"github.com/tetratelabs/wazero/api"
)

type terminationType int

// List of termination reasons
const (
	TerminateFinish = iota
	TerminateRevert
	TerminateSuicide
	TerminateInvalid
)

type WASMInterpreter struct {
	evm *EVM
	cfg Config

	vm *wasmedge.VM
	Contract *Contract

	returnData []byte // Last CALL's return data for subsequent reuse
	
	TxContext
	// StateDB gives access to the underlying state
	StateDB StateDB
	// Depth is the current call stack
	depth int

	Config Config
}



func NewWASMInterpreter(evm *EVM, cfg Config) *WASMInterpreter {

	inter := WASMInterpreter{
		StateDB:  evm.StateDB,
		evm: 	evm,
	}

	return &inter
}



func (in *WASMInterpreter) Run(contract *Contract, input []byte, readOnly bool) (ret []byte, err error) {
	// Increment the call depth which is restricted to 1024
	in.evm.depth++
	defer func() { in.evm.depth-- }()

	in.Contract = contract
	in.Contract.Input = input

	// Create VM with the configure.
	vm := InstantiateWASMEdgeVM()

	in.vm = vm

	vm.LoadWasmBuffer(contract.Code)

	err = vm.Validate()
	if err != nil {
		log.Panicln("🔴 Error while validating WASM module: ", err)
	}

	// Instantiate the WASM module.
	err = vm.Instantiate()
	if err != nil {
	fmt.Println("Instantiation FAILED:", err.Error())
	return
	}


	vm.Release()


	return in.returnData, nil
}

// WriteWASMInterpreter writes the WASMInterpreter to the WASM module's memory.
func WriteWASMInterpreter(module api.Module, in *WASMInterpreter) {
	interpreterBytes, err := json.Marshal(in)
	if err != nil {
		log.Panicln("🔴 Error while marshalling interpreter: ", err)
	}
	
	// The pointer is a linear memory offset, which is where we write the name.
	fmt.Println("interpreterBytes:", len(interpreterBytes))
	

	var lenBytes [4]byte
	binary.LittleEndian.PutUint32(lenBytes[:], uint32(len(interpreterBytes)))

	fmt.Println("lenBytes:", lenBytes)
	if !module.Memory().Write(0, lenBytes[:]) {
		log.Panicf("🟥 Memory.Write(%d, %d) out of range of memory size %d",
			0, len(interpreterBytes), module.Memory().Size())
	}

	if !module.Memory().Write(4, interpreterBytes) {
		log.Panicf("🟥 Memory.Write(%d, %d) out of range of memory size %d",
			0, interpreterBytes, module.Memory().Size())
	}
}