package main

import (
	"flag"
	"fmt"
	addressCli "github.com/GarinAG/gofias/domain/address/delivery/cli"
	versionCli "github.com/GarinAG/gofias/domain/version/delivery/cli"
	indexCli "github.com/GarinAG/gofias/infrastructure/persistence/address/elastic/delivery/cli"
	cli2 "github.com/GarinAG/gofias/infrastructure/persistence/cli"
	"github.com/GarinAG/gofias/infrastructure/registry"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/GarinAG/gofias/util"
	"runtime"
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	ctn, err := registry.NewContainer("cli")
	if err != nil {
		panic(fmt.Sprintf("Failed to init container: %v", err))
	}

	util.CanPrintProcess = ctn.Resolve("config").(interfaces.ConfigInterface).GetBool("process.print")
	app := cli2.NewApp(ctn)

	addressCli.RegisterImportCliEndpoint(app)
	indexCli.RegisterIndexCliEndpoint(app)
	versionCli.RegisterVersionCliEndpoint(app)

	if err := app.Run(); err != nil {
		app.Logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Program fatal error")
	}
}
