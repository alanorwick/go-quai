#!/bin/bash
tinygo build -o hello.wasm -scheduler=none -panic=trap --no-debug -target wasi ./hello.go