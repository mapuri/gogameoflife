# Overview

gogameoflife is a fun project of mine that implements [Conway's Game of Life](https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life) using Golang + WASM

![gogameoflife in action](/snapshots/sc1.png?raw=true "gogameoflife in action")

# Prerequisite
- Golang 1.15
- Depends on syscall/js package which is still in alpha and can have breaking changes across releases of Go

# Compile and run

- Clone the workspace 
- Compile the WASM
```
GOOS=js GOARCH=wasm go build -o resources/game.wasm
```
- Run the webserver
```
go run ./server/server.go -listen=:8080
```
- Point your browser to http://localhost:8080 and enjoy the game of life :)