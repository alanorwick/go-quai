package vm

type terminationType int

// List of termination reasons
const (
	TerminateFinish = iota
	TerminateRevert
	TerminateSuicide
	TerminateInvalid
)

type WASMInterpreter struct {
	EVM *EVM
	cfg Config

	vm       WasmVM
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
		StateDB: evm.StateDB,
		EVM:     evm,
	}

	return &inter
}

func (in *WASMInterpreter) Run(contract *Contract, input []byte, readOnly bool) (ret []byte, err error) {
	// Increment the call depth which is restricted to 1024
	in.EVM.depth++
	defer func() { in.EVM.depth-- }()

	in.Contract = contract
	in.Contract.Input = input

	vm := InstantiateWasmVM(in, in.Contract.Code)
	in.vm = vm

	run, err := in.vm.instance.Exports.GetFunction("run")
	if err != nil {
		panic(err)
	}
	run()

	return in.returnData, nil
}
