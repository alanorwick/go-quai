package vm

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bytecodealliance/wasmtime-go/v7"
	"github.com/tetratelabs/wazero/api"
)


const (
	// EEICallSuccess is the return value in case of a successful contract execution
	EEICallSuccess = 0
	// ErrEEICallFailure is the return value in case of a contract execution failture
	ErrEEICallFailure = 1
	// ErrEEICallRevert is the return value in case a contract calls `revert`
	ErrEEICallRevert = 2
)

const (
	defaultTimeout   = 5 * time.Second
	FuncAbort        = "abort"
	FuncFdWrite      = "fd_write"
	FuncHostStateGet = "hostStateGet"
	FuncHostStateSet = "hostStateSet"
	ModuleEnv        = "env"
	ModuleWasi1      = "wasi_unstable"
	ModuleWasi2      = "wasi_snapshot_preview1"
	ModuleWasmLib    = "WasmLib"
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

var eeiFunctionList = []string{
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
	evm *EVM

	engine     *wasmtime.Engine
	instance   *wasmtime.Instance
	linker     *wasmtime.Linker
	memory     *wasmtime.Memory
	module     *wasmtime.Module
	store      *wasmtime.Store

	Contract *Contract

	cachedResult   []byte
	panicErr       error
	timeoutStarted bool
}


func InstantiateWASMVM() *WasmVM {	
	config := wasmtime.NewConfig()
	// no need to be interruptable by WasmVMBase
	// config.SetInterruptable(true)
	config.SetConsumeFuel(true)
	vm := &WasmVM{engine: wasmtime.NewEngineWithConfig(config)}
	// prevent WasmVMBase from starting timeout interrupting,
	// instead we simply let WasmTime run out of fuel
	vm.timeoutStarted = true // DisableWasmTimeout

	vm.LinkHost()

	return vm
}

func (vm *WasmVM) LinkHost() (err error) {
	vm.store = wasmtime.NewStore(vm.engine)
	vm.linker = wasmtime.NewLinker(vm.engine)

	// new Wasm VM interface
	err = vm.linker.DefineFunc(vm.store, ModuleWasmLib, FuncHostStateGet, vm.HostStateGet)
	if err != nil {
		return err
	}

	return nil
}

func (vm *WasmVM) LoadWasm(wasmData []byte) (err error) {
	vm.module, err = wasmtime.NewModule(vm.engine, wasmData)
	return err
}


func (vm *WasmVM) UnsafeMemory() []byte {
	return vm.memory.UnsafeData(vm.store)
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

func readSize(ctx context.Context, module api.Module, offset uint32, size uint32) []byte {
	buf, ok := module.Memory().Read(offset, size)
	if !ok {
		log.Panicf("🟥 Memory.Read(%d, %d) out of range", 4, size)
	}

	return buf
}


func useGas(ctx context.Context, module api.Module, amount int64) {
	in.gasAccounting(uint64(amount))
}
