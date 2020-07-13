package main

import (
	"fmt"
	addressCli "github.com/GarinAG/gofias/domain/address/delivery/cli"
	versionCli "github.com/GarinAG/gofias/domain/version/delivery/cli"
	indexCli "github.com/GarinAG/gofias/infrastructure/persistence/address/elastic/delivery/cli"
	"github.com/GarinAG/gofias/infrastructure/persistence/config"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/GarinAG/gofias/server/cli"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	appConfig := config.YamlConfig{ConfigPath: "../"}
	err := appConfig.Init()
	if err != nil {
		panic(fmt.Sprintf("Failed to init configuration: %v", err))
	}
	app := cli.NewApp(&appConfig)

	addressCli.RegisterImportCliEndpoint(app)
	versionCli.RegisterVersionCliEndpoint(app)
	indexCli.RegisterIndexCliEndpoint(app)

	if err := app.Run(); err != nil {
		app.Logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Program fatal error")
	}
}
