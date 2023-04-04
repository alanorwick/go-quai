package vm

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"

	"github.com/second-state/WasmEdge-go/wasmedge"
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

func InstantiateWASMEdgeVM() *wasmedge.VM {	
	// Choose the context to use for function calls.
	// Set the logging level.
	wasmedge.SetLogErrorLevel()

	// Create the configure context and add the WASI support.
	// This step is not necessary unless you need WASI support.
	conf := wasmedge.NewConfigure(wasmedge.WASI)
	// Create VM with the configure.
	vm := wasmedge.NewVMWithConfig(conf)

	// Create the module instance with the module name "extern".
	impmod := wasmedge.NewModule("extern")

	// Create and add a function instance into the module instance with export name "func-add".
	functype := wasmedge.NewFunctionType([]wasmedge.ValType{wasmedge.ValType_I32}, []wasmedge.ValType{})
	hostfunc := wasmedge.NewFunction(functype, host_trap, nil, 0)
	functype.Release()
	impmod.AddFunction("trap", hostfunc)

  	// Register the module instance into VM.
	vm.RegisterModule(impmod)

	return vm
}


// Host function body definition.
func host_trap(data interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// add: i32, i32 -> i32
	res := params[0].(int32) + params[1].(int32)
  
	// Set the returns
	returns := make([]interface{}, 1)
	returns[0] = res
  
	// Return
	return returns, wasmedge.Result_Success
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
    in := ReadWASMInterpreter(ctx, module)
	in.gasAccounting(uint64(amount))
	WriteWASMInterpreter(module, in)
}

func ReadWASMInterpreter(ctx context.Context, module api.Module) *WASMInterpreter {
	bufSize, ok := module.Memory().Read(0, 4)
	if !ok {
		log.Panicf("🟥 Memory.Read(%d, %d) out of range", 0, 4)
	}

	size := binary.LittleEndian.Uint32(bufSize)
	
	buf, ok := module.Memory().Read(4, uint32(size))
	if !ok {
		log.Panicf("🟥 Memory.Read(%d, %d) out of range", 4, size)
	}

	in := &WASMInterpreter{}
	json.Unmarshal(buf, in)
	return in
}

func WriteResult(module api.Module, result []byte, resultOffset uint32) {
	module.Memory().Write(resultOffset, result)
}

