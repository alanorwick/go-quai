#!/bin/bash
tinygo build -o get_address.wasm -scheduler=none -panic=trap --no-debug -target wasi ./get_address.go