package vm

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
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

func InstantiateWASMRuntime(ctx context.Context) wazero.Runtime {

	// Create a new WebAssembly Runtime.
	r := wazero.NewRuntime(ctx)

	_, errEnv := r.NewHostModuleBuilder("env").
	NewFunctionBuilder().WithFunc(logUint32).Export("hostLogUint32").
	NewFunctionBuilder().WithFunc(logString).Export("hostLogString").
	NewFunctionBuilder().WithFunc(useGas).Export("useGas").
	Instantiate(ctx)
	if errEnv != nil {
		fmt.Println("🔴 Error with env:", errEnv)
	}

	_, errInstantiate := wasi_snapshot_preview1.Instantiate(ctx, r)
	if errInstantiate != nil {
		fmt.Println("🔴 Error with Instantiate:", errInstantiate)
	}

	
	fmt.Println("🤖: Instantiated WASM Runtime")


	return r
}


func logUint32(value uint32) {
	fmt.Println("🤖:", value)
}

func logString(ctx context.Context, module api.Module, offset, byteCount uint32) {
	buf, ok := module.Memory().Read(offset, byteCount)
	if !ok {
		log.Panicf("🟥 Memory.Read(%d, %d) out of range", offset, byteCount)
	}
	fmt.Println("👽:", string(buf))
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


func useGas(ctx context.Context, module api.Module, amount int64) {
    in := ReadWASMInterpreter(module)
	in.gasAccounting(uint64(amount))
	WriteWASMInterpreter(module, in)
}

func ReadWASMInterpreter(module api.Module) *WASMInterpreter {
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