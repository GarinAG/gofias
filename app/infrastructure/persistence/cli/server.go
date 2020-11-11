package cli

import (
	"github.com/GarinAG/gofias/domain/address/service"
	directoryService "github.com/GarinAG/gofias/domain/directory/service"
	fiasApiService "github.com/GarinAG/gofias/domain/fiasApi/service"
	osmService "github.com/GarinAG/gofias/domain/osm/service"
	versionService "github.com/GarinAG/gofias/domain/version/service"
	"github.com/GarinAG/gofias/infrastructure/registry"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/urfave/cli/v2"
	"os"
)

// Объект приложения
type App struct {
	Server           *cli.App                           // CLI сервер приложения
	Container        *registry.Container                // Контейнер зависимостей
	Config           interfaces.ConfigInterface         // Конфигурации
	Logger           interfaces.LoggerInterface         // Логгер
	ImportService    *service.ImportService             // Сервис импорта
	AddressService   *service.AddressImportService      // Сервис импорта адресов
	HouseService     *service.HouseImportService        // Сервис импорта домов
	VersionService   *versionService.VersionService     // Сервис версий
	DirectoryService *directoryService.DirectoryService // Сервис управления файлами
	FiasApiService   *fiasApiService.FiasApiService     // Сервис ФИАС
	OsmService       *osmService.OsmService             // Сервис OSM
}

// Инициализация приложения
func NewApp(ctn *registry.Container) *App {
	// Инициализация сервера
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
		HouseService:     ctn.Resolve("houseImportService").(*service.HouseImportService),
		VersionService:   ctn.Resolve("versionService").(*versionService.VersionService),
		FiasApiService:   ctn.Resolve("fiasApiService").(*fiasApiService.FiasApiService),
		OsmService:       ctn.Resolve("osmService").(*osmService.OsmService),
	}
}

// Инициализация CLI сервера
func initCli() *cli.App {
	app := cli.App{
		Name:    "fiascli",
		Usage:   "Cli fias program",
		Version: "0.1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config-path",
				Value: "./",
				Usage: "Config path",
			},
			&cli.StringFlag{
				Name:  "config-type",
				Value: "yaml",
				Usage: "Config type",
			},
		},
	}

	return &app
}

// Запуск сервера
func (a *App) Run() error {
	err := a.Server.Run(os.Args)
	return err
}
