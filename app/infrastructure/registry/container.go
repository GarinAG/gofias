package registry

import (
	importService "github.com/GarinAG/gofias/application"
	"github.com/GarinAG/gofias/domain/address/service"
	directoryService "github.com/GarinAG/gofias/domain/directory/service"
	fiasApiService "github.com/GarinAG/gofias/domain/fiasApi/service"
	versionService "github.com/GarinAG/gofias/domain/version/service"
	elasticRepository "github.com/GarinAG/gofias/infrastructure/persistence/address/elastic/repository"
	helper "github.com/GarinAG/gofias/infrastructure/persistence/elastic"
	fiasApiRepository "github.com/GarinAG/gofias/infrastructure/persistence/fiasApi/http/repository"
	log "github.com/GarinAG/gofias/infrastructure/persistence/logger"
	versionRepository "github.com/GarinAG/gofias/infrastructure/persistence/version/elastic/repository"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/olivere/elastic/v7"
	"github.com/sarulabs/di"
)

type Container struct {
	ctn di.Container
}

func NewContainer(config interfaces.ConfigInterface) (*Container, error) {
	builder, err := di.NewBuilder()
	if err != nil {
		return nil, err
	}

	if err := builder.Add([]di.Def{
		{
			Name: "logger",
			Build: func(ctn di.Container) (interface{}, error) {
				logger := log.InitLogrusLogger(config)
				return &logger, nil
			},
		},
		{
			Name: "directoryService",
			Build: func(ctn di.Container) (interface{}, error) {
				return directoryService.NewDirectoryService(ctn.Get("logger").(interfaces.LoggerInterface), config), nil
			},
		},
		{
			Name: "elasticClient",
			Build: func(ctn di.Container) (interface{}, error) {
				return helper.InitElasticClient(config), nil
			},
		},
		{
			Name: "addressService",
			Build: func(ctn di.Container) (interface{}, error) {
				repo := elasticRepository.NewElasticAddressRepository(ctn.Get("elasticClient").(*elastic.Client), config)
				return service.NewAddressService(repo, ctn.Get("logger").(interfaces.LoggerInterface)), nil
			},
		},
		{
			Name: "houseService",
			Build: func(ctn di.Container) (interface{}, error) {
				repo := elasticRepository.NewElasticHouseRepository(ctn.Get("elasticClient").(*elastic.Client), config)
				return service.NewHouseService(repo, ctn.Get("logger").(interfaces.LoggerInterface)), nil
			},
		},
		{
			Name: "versionService",
			Build: func(ctn di.Container) (interface{}, error) {
				repo := versionRepository.NewElasticVersionRepository(ctn.Get("elasticClient").(*elastic.Client), config)
				return versionService.NewVersionService(repo, ctn.Get("logger").(interfaces.LoggerInterface)), nil
			},
		},
		{
			Name: "fiasApiService",
			Build: func(ctn di.Container) (interface{}, error) {
				repo := fiasApiRepository.NewHttpFiasApiRepository(config)
				return fiasApiService.NewFiasApiService(repo, ctn.Get("logger").(interfaces.LoggerInterface)), nil
			},
		},
		{
			Name: "importService",
			Build: func(ctn di.Container) (interface{}, error) {
				return importService.NewImportService(
					ctn.Get("logger").(interfaces.LoggerInterface),
					ctn.Get("directoryService").(*directoryService.DirectoryService),
					ctn.Get("addressService").(*service.AddressImportService),
					ctn.Get("houseService").(*service.HouseImportService)), nil
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
