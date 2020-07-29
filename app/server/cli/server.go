package cli

import (
	"github.com/GarinAG/gofias/domain/address/service"
	directoryService "github.com/GarinAG/gofias/domain/directory/service"
	fiasApiService "github.com/GarinAG/gofias/domain/fiasApi/service"
	versionService "github.com/GarinAG/gofias/domain/version/service"
	"github.com/GarinAG/gofias/infrastructure/registry"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/urfave/cli/v2"
	"os"
)

type App struct {
	Server           *cli.App
	Container        *registry.Container
	Config           interfaces.ConfigInterface
	Logger           interfaces.LoggerInterface
	ImportService    *service.ImportService
	AddressService   *service.AddressImportService
	HouseService     *service.HouseImportService
	VersionService   *versionService.VersionService
	DirectoryService *directoryService.DirectoryService
	FiasApiService   *fiasApiService.FiasApiService
}

func NewApp(ctn *registry.Container) *App {
	server := initCli()
	logger := ctn.Resolve("logger").(interfaces.LoggerInterface)

	defer func() {
		if r := recover(); r != nil {
			logger.WithFields(interfaces.LoggerFields{"error": r}).Panic("Program fatal error")
			os.Exit(1)
		}
	}()

	return &App{
		Server:           server,
		Container:        ctn,
		Config:           ctn.Resolve("config").(interfaces.ConfigInterface),
		Logger:           logger,
		DirectoryService: ctn.Resolve("directoryService").(*directoryService.DirectoryService),
		ImportService:    ctn.Resolve("importService").(*service.ImportService),
		AddressService:   ctn.Resolve("addressImportService").(*service.AddressImportService),
		HouseService:     ctn.Resolve("houseService").(*service.HouseImportService),
		VersionService:   ctn.Resolve("versionService").(*versionService.VersionService),
		FiasApiService:   ctn.Resolve("fiasApiService").(*fiasApiService.FiasApiService),
	}
}

func initCli() *cli.App {
	app := cli.App{
		Name:    "fiascli",
		Usage:   "Cli fias program",
		Version: "0.1.0",
	}

	return &app
}

func (a *App) Run() error {
	err := a.Server.Run(os.Args)
	return err
}
