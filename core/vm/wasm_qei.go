package vm

import (
	"fmt"

	"github.com/wasmerio/wasmer-go/wasmer"
)

const (
	// QEICallSuccess is the return value in case of a successful contract execution
	QEICallSuccess = 0
	// ErrQEICallFailure is the return value in case of a contract execution failture
	ErrQEICallFailure = 1
	// ErrQEICallRevert is the return value in case a contract calls `revert`
	ErrQEICallRevert = 2
)

// List of gas costs
const (
	GasCostZero           = 0
	GasCostBase           = 2
	GasCostVeryLow        = 3
	GasCostLow            = 5
	GasCostMid            = 8
	GasCostHigh           = 10
	GasCostExtCode        = 700
	GasCostBalance        = 400
	GasCostSLoad          = 200
	GasCostJumpDest       = 1
	GasCostSSet           = 20000
	GasCostSReset         = 5000
	GasRefundSClear       = 15000
	GasRefundSelfDestruct = 24000
	GasCostCreate         = 32000
	GasCostCall           = 700
	GasCostCallValue      = 9000
	GasCostCallStipend    = 2300
	GasCostNewAccount     = 25000
	GasCostLog            = 375
	GasCostLogData        = 8
	GasCostLogTopic       = 375
	GasCostCopy           = 3
	GasCostBlockHash      = 800
)

var QEIFunctionList = []string{
	"useGas",
	"getAddress",
	"getExternalBalance",
	"getBlockHash",
	"call",
	"callDataCopy",
	"getCallDataSize",
	"callCode",
	"callDelegate",
	"callStatic",
	"storageStore",
	"storageLoad",
	"getCaller",
	"getCallValue",
	"codeCopy",
	"getCodeSize",
	"getBlockCoinbase",
	"create",
	"getBlockDifficulty",
	"externalCodeCopy",
	"getExternalCodeSize",
	"getGasLeft",
	"getBlockGasLimit",
	"getTxGasPrice",
	"log",
	"getBlockNumber",
	"getTxOrigin",
	"finish",
	"revert",
	"getReturnDataSize",
	"returnDataCopy",
	"selfDestruct",
	"getBlockTimestamp",
}

type WasmVM struct {
	engine   *wasmer.Engine
	store    *wasmer.Store
	imports  *wasmer.ImportObject
	module   *wasmer.Module
	instance *wasmer.Instance
	test     string
}

func InstantiateWasmVM(in *WASMInterpreter, code []byte) WasmVM {

	// Create a new WebAssembly Runtime.
	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)

	vm := WasmVM{
		engine: engine,
		store:  store,
	}

	fmt.Println("🤖: Instantiated WASM Runtime")

	// Create the new WASM module
	module, _ := wasmer.NewModule(store, code)

	wasiEnv, _ := wasmer.NewWasiStateBuilder("wasi-program").
		// Choose according to your actual situation
		// Argument("--foo").
		// Environment("ABC", "DEF").
		// MapDirectory("./", ".").
		Finalize()

	// Let's use the new `ImportObject` API…
	importObject, err := wasiEnv.GenerateImportObject(store, module)
	if err != nil {
		panic(err)
	}

	useGasFunc := wasmer.NewFunctionWithEnvironment(
		store,
		wasmer.NewFunctionType(wasmer.NewValueTypes(wasmer.I64), wasmer.NewValueTypes()),
		in,
		useGas,
	)

	getAddressFunc := wasmer.NewFunctionWithEnvironment(
		store,
		wasmer.NewFunctionType(wasmer.NewValueTypes(wasmer.I32), wasmer.NewValueTypes()),
		in,
		vm.getAddress,
	)

	// … to register the `math.sum` function.
	importObject.Register(
		"env",
		map[string]wasmer.IntoExtern{
			"useGas":     useGasFunc,
			"getAddress": getAddressFunc,
		},
	)

	// Instantiate the module.
	instance, err := wasmer.NewInstance(module, importObject)
	if err != nil {
		panic(err)
	}

	vm.imports = importObject
	vm.module = module
	vm.test = "test"
	vm.instance = instance

	return vm
}

func (in WASMInterpreter) gasAccounting(cost uint64) {
	if in.Contract == nil {
		panic("nil contract")
	}
	if cost > in.Contract.Gas {
		panic(fmt.Sprintf("out of gas %d > %d", cost, in.Contract.Gas))
	}
	in.Contract.Gas -= cost
}

func useGas(environment interface{}, args []wasmer.Value) ([]wasmer.Value, error) {
	in := environment.(*WASMInterpreter)
	amount := args[0].I64()
	fmt.Println("🤖: useGas", amount)
	fmt.Println("🤖: useGas", in.Contract.Gas)
	fmt.Println("🤖: useGas", in)
	in.gasAccounting(uint64(amount))
	return []wasmer.Value{}, nil
}

func (vm *WasmVM) getAddress(environment interface{}, args []wasmer.Value) ([]wasmer.Value, error) {
	in := environment.(*WASMInterpreter)
	resultOffset := args[0].I32()
	in.gasAccounting(GasCostBase)
	fmt.Println("🤖: getAddress", in.Contract.CodeAddr)
	addr := in.Contract.CodeAddr.Bytes()

	setAt, err := vm.instance.Exports.GetFunction("set_at")
	if err != nil {
		panic(fmt.Sprintln("Failed to retrieve the `set_at` function:", err))
	}

	_, err = setAt(resultOffset, addr)
	if err != nil {
		panic(fmt.Sprintln("Failed to call the `set_at` function:", err))
	}
	return []wasmer.Value{}, nil
}
