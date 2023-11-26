#!/bin/bash
tinygo build -o wrc-20.wasm -scheduler=none --no-debug -target wasi ./wrc-20.go