package registry

import (
	"flag"
	"github.com/GarinAG/gofias/domain/address/repository"
	"github.com/GarinAG/gofias/domain/address/service"
	directoryService "github.com/GarinAG/gofias/domain/directory/service"
	fiasApiService "github.com/GarinAG/gofias/domain/fiasApi/service"
	osmService "github.com/GarinAG/gofias/domain/osm/service"
	versionService "github.com/GarinAG/gofias/domain/version/service"
	elasticRepository "github.com/GarinAG/gofias/infrastructure/persistence/address/elastic/repository"
	"github.com/GarinAG/gofias/infrastructure/persistence/config"
	elasticHelper "github.com/GarinAG/gofias/infrastructure/persistence/elastic"
	fiasApiRepository "github.com/GarinAG/gofias/infrastructure/persistence/fiasApi/http/repository"
	log "github.com/GarinAG/gofias/infrastructure/persistence/logger"
	versionRepository "github.com/GarinAG/gofias/infrastructure/persistence/version/elastic/repository"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/sarulabs/di"
)

var (
	ConfigPath = flag.String("config-path", "./", "Config path")
	ConfigType = flag.String("config-type", "yaml", "Config type")
)

type Container struct {
	ctn di.Container
}

func NewContainer(loggerPrefix string) (*Container, error) {
	builder, err := di.NewBuilder()
	if err != nil {
		return nil, err
	}

	if err := builder.Add([]di.Def{
		{
			Name: "config",
			Build: func(ctn di.Container) (interface{}, error) {
				appConfig := config.ViperConfig{ConfigPath: *ConfigPath, ConfigType: *ConfigType}
				err := appConfig.Init()

				return &appConfig, err
			},
		},
		{
			Name: "logger",
			Build: func(ctn di.Container) (interface{}, error) {
				appConfig := ctn.Get("config").(interfaces.ConfigInterface)
				loggerConfig := interfaces.LoggerConfiguration{
					EnableConsole:      appConfig.GetBool("logger.console.enable"),
					ConsoleLevel:       appConfig.GetString("logger.console.level"),
					ConsoleJSONFormat:  appConfig.GetBool("logger.console.json"),
					EnableFile:         appConfig.GetBool("logger.file.enable"),
					FileLevel:          appConfig.GetString("logger.file.level"),
					FileJSONFormat:     appConfig.GetBool("logger.file.json"),
					FileLocation:       appConfig.GetString("logger.file.path"),
					FileLocationPrefix: loggerPrefix,
				}
				logger := log.NewZapLogger(loggerConfig)

				return logger, nil
			},
		},
		{
			Name: "elasticClient",
			Build: func(ctn di.Container) (interface{}, error) {
				client := elasticHelper.NewElasticClient(ctn.Get("config").(interfaces.ConfigInterface), ctn.Get("logger").(interfaces.LoggerInterface))

				return client, nil
			},
		},
		{
			Name: "houseRepository",
			Build: func(ctn di.Container) (interface{}, error) {
				appConfig := ctn.Get("config").(interfaces.ConfigInterface)
				repo := elasticRepository.NewElasticHouseRepository(
					ctn.Get("elasticClient").(*elasticHelper.Client),
					ctn.Get("logger").(interfaces.LoggerInterface),
					appConfig.GetInt("batch.size"),
					appConfig.GetString("project.prefix"),
					appConfig.GetInt("workers.houses"))

				return repo, nil
			},
		},
		{
			Name: "addressRepository",
			Build: func(ctn di.Container) (interface{}, error) {
				appConfig := ctn.Get("config").(interfaces.ConfigInterface)
				repo := elasticRepository.NewElasticAddressRepository(
					ctn.Get("elasticClient").(*elasticHelper.Client),
					ctn.Get("logger").(interfaces.LoggerInterface),
					appConfig.GetInt("batch.size"),
					appConfig.GetString("project.prefix"),
					appConfig.GetInt("workers.addresses"))

				return repo, nil
			},
		},
		{
			Name: "downloadService",
			Build: func(ctn di.Container) (interface{}, error) {
				return directoryService.NewDownloadService(
					ctn.Get("logger").(interfaces.LoggerInterface),
					ctn.Get("config").(interfaces.ConfigInterface)), nil
			},
		},
		{
			Name: "directoryService",
			Build: func(ctn di.Container) (interface{}, error) {
				return directoryService.NewDirectoryService(
					ctn.Get("downloadService").(*directoryService.DownloadService),
					ctn.Get("logger").(interfaces.LoggerInterface),
					ctn.Get("config").(interfaces.ConfigInterface)), nil
			},
		},
		{
			Name: "addressImportService",
			Build: func(ctn di.Container) (interface{}, error) {
				repo := ctn.Get("addressRepository").(repository.AddressRepositoryInterface)
				logger := ctn.Get("logger").(interfaces.LoggerInterface)

				return service.NewAddressImportService(repo, logger), nil
			},
		},
		{
			Name: "houseImportService",
			Build: func(ctn di.Container) (interface{}, error) {
				repo := ctn.Get("houseRepository").(repository.HouseRepositoryInterface)
				logger := ctn.Get("logger").(interfaces.LoggerInterface)

				return service.NewHouseImportService(repo, logger), nil
			},
		},
		{
			Name: "versionService",
			Build: func(ctn di.Container) (interface{}, error) {
				repo := versionRepository.NewElasticVersionRepository(ctn.Get("elasticClient").(*elasticHelper.Client),
					ctn.Get("config").(interfaces.ConfigInterface))
				return versionService.NewVersionService(repo, ctn.Get("logger").(interfaces.LoggerInterface)), nil
			},
		},
		{
			Name: "fiasApiService",
			Build: func(ctn di.Container) (interface{}, error) {
				repo := fiasApiRepository.NewHttpFiasApiRepository(ctn.Get("config").(interfaces.ConfigInterface))
				return fiasApiService.NewFiasApiService(repo, ctn.Get("logger").(interfaces.LoggerInterface)), nil
			},
		},
		{
			Name: "importService",
			Build: func(ctn di.Container) (interface{}, error) {
				return service.NewImportService(
					ctn.Get("logger").(interfaces.LoggerInterface),
					ctn.Get("directoryService").(*directoryService.DirectoryService),
					ctn.Get("addressImportService").(*service.AddressImportService),
					ctn.Get("houseImportService").(*service.HouseImportService),
					ctn.Get("config").(interfaces.ConfigInterface)), nil
			},
		},
		{
			Name: "addressService",
			Build: func(ctn di.Container) (interface{}, error) {
				repo := ctn.Get("addressRepository").(repository.AddressRepositoryInterface)
				logger := ctn.Get("logger").(interfaces.LoggerInterface)

				return service.NewAddressService(repo, logger), nil
			},
		},
		{
			Name: "houseService",
			Build: func(ctn di.Container) (interface{}, error) {
				repo := ctn.Get("houseRepository").(repository.HouseRepositoryInterface)
				logger := ctn.Get("logger").(interfaces.LoggerInterface)

				return service.NewHouseService(repo, logger), nil
			},
		},
		{
			Name: "osmService",
			Build: func(ctn di.Container) (interface{}, error) {
				addressRepo := ctn.Get("addressRepository").(repository.AddressRepositoryInterface)
				houseRepo := ctn.Get("houseRepository").(repository.HouseRepositoryInterface)
				logger := ctn.Get("logger").(interfaces.LoggerInterface)
				downloadService := ctn.Get("downloadService").(*directoryService.DownloadService)
				appConfig := ctn.Get("config").(interfaces.ConfigInterface)

				return osmService.NewOsmService(addressRepo, houseRepo, downloadService, logger, appConfig), nil
			},
		},
	}...); err != nil {
		return nil, err
	}

	return &Container{
		ctn: builder.Build(),
	}, nil
}

func (c *Container) Resolve(name string) interface{} {
	return c.ctn.Get(name)
}

func (c *Container) Clean() error {
	return c.ctn.Clean()
}
