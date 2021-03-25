package registry

import (
	"flag"
	cache "github.com/AeroAgency/golang-bigcache-lib"
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
	"github.com/allegro/bigcache"
	"github.com/sarulabs/di"
	"time"
)

var (
	ConfigPath   = flag.String("config-path", "./", "Config path")
	ConfigType   = flag.String("config-type", "yaml", "Config type")
	LoggerPrefix = flag.String("logger-prefix", "cli", "Logger prefix")
)

// Объект контейнера зависимостей
type Container struct {
	ctn di.Container // Контейнер
}

// Инициализация контейнера
func NewContainer() (*Container, error) {
	builder, err := di.NewBuilder()
	if err != nil {
		return nil, err
	}

	if err := builder.Add([]di.Def{
		// Конфигурация
		{
			Name: "config",
			Build: func(ctn di.Container) (interface{}, error) {
				appConfig := config.ViperConfig{ConfigPath: *ConfigPath, ConfigType: *ConfigType}
				err := appConfig.Init()

				return &appConfig, err
			},
		},
		// Логгер
		{
			Name: "logger",
			Build: func(ctn di.Container) (interface{}, error) {
				appConfig := ctn.Get("config").(interfaces.ConfigInterface)
				loggerConfig := interfaces.LoggerConfiguration{
					EnableConsole:      appConfig.GetConfig().LoggerConsole.Enable,
					ConsoleLevel:       appConfig.GetConfig().LoggerConsole.Level,
					ConsoleJSONFormat:  appConfig.GetConfig().LoggerConsole.Json,
					EnableFile:         appConfig.GetConfig().LoggerFile.Enable,
					FileLevel:          appConfig.GetConfig().LoggerFile.Level,
					FileJSONFormat:     appConfig.GetConfig().LoggerFile.Json,
					FileLocation:       appConfig.GetConfig().LoggerFile.Path,
					FileLocationPrefix: *LoggerPrefix,
				}
				logger := log.NewZapLogger(loggerConfig)

				return logger, nil
			},
		},
		// Кэш
		{
			Name: "cache",
			Build: func(ctn di.Container) (interface{}, error) {
				cacheConfig := bigcache.Config{
					Shards:             1024,
					LifeWindow:         10 * time.Minute,
					CleanWindow:        5 * time.Minute,
					MaxEntriesInWindow: 1000 * 10 * 60,
					MaxEntrySize:       500,
					Verbose:            false,
					HardMaxCacheSize:   2048,
					OnRemove:           nil,
					OnRemoveWithReason: nil,
				}
				bigCacheInstance, _ := bigcache.NewBigCache(cacheConfig)
				cacheInstance := cache.NewBigCache(bigCacheInstance)

				return cacheInstance, nil
			},
		},
		// Клиент эластика
		{
			Name: "elasticClient",
			Build: func(ctn di.Container) (interface{}, error) {
				client := elasticHelper.NewElasticClient(ctn.Get("config").(interfaces.ConfigInterface), ctn.Get("logger").(interfaces.LoggerInterface))

				return client, nil
			},
		},
		// Репозиторий домов
		{
			Name: "houseRepository",
			Build: func(ctn di.Container) (interface{}, error) {
				appConfig := ctn.Get("config").(interfaces.ConfigInterface)
				repo := elasticRepository.NewElasticHouseRepository(
					ctn.Get("elasticClient").(*elasticHelper.Client),
					ctn.Get("logger").(interfaces.LoggerInterface),
					appConfig.GetConfig().BatchSize,
					appConfig.GetConfig().ProjectPrefix,
					appConfig.GetConfig().Workers.Houses)

				return repo, nil
			},
		},
		// Репозиторий адресов
		{
			Name: "addressRepository",
			Build: func(ctn di.Container) (interface{}, error) {
				appConfig := ctn.Get("config").(interfaces.ConfigInterface)
				repo := elasticRepository.NewElasticAddressRepository(
					ctn.Get("elasticClient").(*elasticHelper.Client),
					ctn.Get("logger").(interfaces.LoggerInterface),
					appConfig.GetConfig().BatchSize,
					appConfig.GetConfig().ProjectPrefix,
					appConfig.GetConfig().Workers.Addresses,
					ctn.Get("cache").(cache.CacheInterface))

				return repo, nil
			},
		},
		// Сервис загрузок
		{
			Name: "downloadService",
			Build: func(ctn di.Container) (interface{}, error) {
				return directoryService.NewDownloadService(
					ctn.Get("logger").(interfaces.LoggerInterface),
					ctn.Get("config").(interfaces.ConfigInterface)), nil
			},
		},
		// Сервис работы с файлами
		{
			Name: "directoryService",
			Build: func(ctn di.Container) (interface{}, error) {
				return directoryService.NewDirectoryService(
					ctn.Get("downloadService").(*directoryService.DownloadService),
					ctn.Get("logger").(interfaces.LoggerInterface),
					ctn.Get("config").(interfaces.ConfigInterface)), nil
			},
		},
		// Сервис импорта адресов
		{
			Name: "addressImportService",
			Build: func(ctn di.Container) (interface{}, error) {
				repo := ctn.Get("addressRepository").(repository.AddressRepositoryInterface)
				logger := ctn.Get("logger").(interfaces.LoggerInterface)

				return service.NewAddressImportService(repo, logger), nil
			},
		},
		// Сервис импорта домов
		{
			Name: "houseImportService",
			Build: func(ctn di.Container) (interface{}, error) {
				repo := ctn.Get("houseRepository").(repository.HouseRepositoryInterface)
				logger := ctn.Get("logger").(interfaces.LoggerInterface)

				return service.NewHouseImportService(repo, logger), nil
			},
		},
		// Сервис версий
		{
			Name: "versionService",
			Build: func(ctn di.Container) (interface{}, error) {
				repo := versionRepository.NewElasticVersionRepository(ctn.Get("elasticClient").(*elasticHelper.Client),
					ctn.Get("config").(interfaces.ConfigInterface))
				return versionService.NewVersionService(repo, ctn.Get("logger").(interfaces.LoggerInterface)), nil
			},
		},
		// Сервис работы с ФИАС API
		{
			Name: "fiasApiService",
			Build: func(ctn di.Container) (interface{}, error) {
				appConfig := ctn.Get("config").(interfaces.ConfigInterface)
				logger := ctn.Get("logger").(interfaces.LoggerInterface)
				repo := fiasApiRepository.NewHttpFiasApiRepository(appConfig)
				return fiasApiService.NewFiasApiService(repo, logger, appConfig), nil
			},
		},
		// Сервис импорта
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
		// Сервис адресов
		{
			Name: "addressService",
			Build: func(ctn di.Container) (interface{}, error) {
				repo := ctn.Get("addressRepository").(repository.AddressRepositoryInterface)
				logger := ctn.Get("logger").(interfaces.LoggerInterface)

				return service.NewAddressService(repo, logger), nil
			},
		},
		// Сервис домов
		{
			Name: "houseService",
			Build: func(ctn di.Container) (interface{}, error) {
				repo := ctn.Get("houseRepository").(repository.HouseRepositoryInterface)
				logger := ctn.Get("logger").(interfaces.LoggerInterface)

				return service.NewHouseService(repo, logger), nil
			},
		},
		// Сервис работы с OpenStreetMap
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

// Получить зависимость
func (c *Container) Resolve(name string) interface{} {
	return c.ctn.Get(name)
}

// Очистить контейнер
func (c *Container) Clean() error {
	return c.ctn.Clean()
}
