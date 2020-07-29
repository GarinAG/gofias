package main

import (
	"fmt"
	"github.com/GarinAG/gofias/infrastructure/registry"
	"github.com/GarinAG/gofias/server/grpc"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	ctn, err := registry.NewContainer()
	if err != nil {
		panic(fmt.Sprintf("Failed to init container: %v", err))
	}

	app := grpc.NewGrpcServer(ctn)

	if err := app.Run(); err != nil {
		app.Logger.Fatal(err.Error())
	}
}
