package main

import (
	"flag"
	"fmt"
	grpc2 "github.com/GarinAG/gofias/infrastructure/persistence/grpc"
	"github.com/GarinAG/gofias/infrastructure/registry"
	"github.com/GarinAG/gofias/interfaces"
	"runtime"
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	ctn, err := registry.NewContainer("grpc")
	if err != nil {
		panic(fmt.Sprintf("Failed to init container: %v", err))
	}

	app := grpc2.NewGrpcServer(ctn)

	if err := app.Run(); err != nil {
		app.Logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Program fatal error")
	}
}
