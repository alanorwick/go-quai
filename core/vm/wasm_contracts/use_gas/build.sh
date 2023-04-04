#!/bin/bash
tinygo build -o use_gas.wasm -scheduler=none -panic=trap --no-debug -target wasi ./use_gas.go