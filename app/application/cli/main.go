package main

import (
	"fmt"
	addressCli "github.com/GarinAG/gofias/domain/address/delivery/cli"
	versionCli "github.com/GarinAG/gofias/domain/version/delivery/cli"
	indexCli "github.com/GarinAG/gofias/infrastructure/persistence/address/elastic/delivery/cli"
	"github.com/GarinAG/gofias/infrastructure/persistence/config"
	log "github.com/GarinAG/gofias/infrastructure/persistence/logger"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/GarinAG/gofias/server/cli"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	appConfig := initConfig()
	logger := initLogger(appConfig)
	app := cli.NewApp(appConfig, logger)

	addressCli.RegisterImportCliEndpoint(app)
	indexCli.RegisterIndexCliEndpoint(app)
	versionCli.RegisterVersionCliEndpoint(app)

	if err := app.Run(); err != nil {
		app.Logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Program fatal error")
	}
}

func initConfig() interfaces.ConfigInterface {
	appConfig := config.YamlConfig{ConfigPath: "../"}
	err := appConfig.Init()
	if err != nil {
		panic(fmt.Sprintf("Failed to init configuration: %v", err))
	}
	return &appConfig
}

func initLogger(config interfaces.ConfigInterface) interfaces.LoggerInterface {
	loggerConfig := interfaces.LoggerConfiguration{
		EnableConsole:     config.GetBool("logger.console.enable"),
		ConsoleLevel:      config.GetString("logger.console.level"),
		ConsoleJSONFormat: config.GetBool("logger.console.json"),
		EnableFile:        config.GetBool("logger.file.enable"),
		FileLevel:         config.GetString("logger.file.level"),
		FileJSONFormat:    config.GetBool("logger.file.json"),
		FileLocation:      config.GetString("logger.file.path"),
	}

	logger, err := log.NewZapLogger(loggerConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to init logger: %v", err))
	}

	return logger
}
