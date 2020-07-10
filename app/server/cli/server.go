package cli

import (
	"fmt"
	importService "github.com/GarinAG/gofias/application"
	"github.com/GarinAG/gofias/domain/address/service"
	directoryService "github.com/GarinAG/gofias/domain/directory/service"
	fiasApiService "github.com/GarinAG/gofias/domain/fiasApi/service"
	versionService "github.com/GarinAG/gofias/domain/version/service"
	"github.com/GarinAG/gofias/infrastructure/persistence/config"
	"github.com/GarinAG/gofias/infrastructure/registry"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/urfave/cli/v2"
	"os"
)

type App struct {
	Server           *cli.App
	Logger           interfaces.LoggerInterface
	ImportService    *importService.ImportService
	AddressService   *service.AddressImportService
	HouseService     *service.HouseImportService
	VersionService   *versionService.VersionService
	DirectoryService *directoryService.DirectoryService
	FiasApiService   *fiasApiService.FiasApiService
}

func NewApp() *App {
	yamlConfig := config.YamlConfig{ConfigPath: "../"}
	//envConfig := config.EnvConfig{}
	err := yamlConfig.Init()
	if err != nil {
		panic(fmt.Sprintf("Failed to init configuration: %v", err))
	}
	ctn, err := registry.NewContainer(&yamlConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to build container: %v", err))
	}
	logger := ctn.Resolve("logger").(interfaces.LoggerInterface)
	server := initCli()
	defer func() {
		if r := recover(); r != nil {
			logger.Panic("Error: ", r)
			os.Exit(1)
		}
	}()

	return &App{
		Server:           server,
		Logger:           logger,
		DirectoryService: ctn.Resolve("directoryService").(*directoryService.DirectoryService),
		ImportService:    ctn.Resolve("importService").(*importService.ImportService),
		AddressService:   ctn.Resolve("addressService").(*service.AddressImportService),
		HouseService:     ctn.Resolve("houseService").(*service.HouseImportService),
		VersionService:   ctn.Resolve("versionService").(*versionService.VersionService),
		FiasApiService:   ctn.Resolve("fiasApiService").(*fiasApiService.FiasApiService),
	}
}

func initCli() *cli.App {
	app := cli.App{
		Name:    "fiascli",
		Usage:   "cli fias program",
		Version: "0.1.0",
	}

	return &app
}

func (a *App) Run() error {
	err := a.Server.Run(os.Args)
	return err
}
