package main

import "syscall/js"

func main() {
	registerCallbacks()
	resetGame(js.Null(), nil)
	println("go wasm initialized")
	// block for ever, so wasm functions stay available
	select {}
}
