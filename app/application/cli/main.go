package main

import (
	"fmt"
	addressCli "github.com/GarinAG/gofias/domain/address/delivery/cli"
	versionCli "github.com/GarinAG/gofias/domain/version/delivery/cli"
	indexCli "github.com/GarinAG/gofias/infrastructure/persistence/address/elastic/delivery/cli"
	"github.com/GarinAG/gofias/infrastructure/registry"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/GarinAG/gofias/server/cli"
	"github.com/GarinAG/gofias/util"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	ctn, err := registry.NewContainer()
	if err != nil {
		panic(fmt.Sprintf("Failed to init container: %v", err))
	}

	util.CanPrintProcess = ctn.Resolve("config").(interfaces.ConfigInterface).GetBool("process.print")
	app := cli.NewApp(ctn)

	addressCli.RegisterImportCliEndpoint(app)
	indexCli.RegisterIndexCliEndpoint(app)
	versionCli.RegisterVersionCliEndpoint(app)

	if err := app.Run(); err != nil {
		app.Logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Program fatal error")
	}
}
