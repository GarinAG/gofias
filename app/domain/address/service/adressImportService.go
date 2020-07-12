package service

import (
	"encoding/xml"
	"errors"
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/domain/address/repository"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/GarinAG/gofias/util"
	"os"
	"sync"
	"time"
)

type AddressImportService struct {
	addressRepo repository.AddressRepositoryInterface
	logger      interfaces.LoggerInterface
}

func NewAddressService(addressRepo repository.AddressRepositoryInterface, logger interfaces.LoggerInterface) *AddressImportService {
	err := addressRepo.Init()
	if err != nil {
		logger.Panic(err.Error())
		os.Exit(1)
	}

	return &AddressImportService{
		addressRepo: addressRepo,
		logger:      logger,
	}
}

func (a *AddressImportService) Import(filePath string, wg *sync.WaitGroup, cnt chan int) {
	defer wg.Done()
	addressChannel := make(chan interface{})
	done := make(chan bool)
	go util.ParseFile(filePath, addressChannel, done, a.logger, a.ParseElement)
	go a.addressRepo.InsertUpdateCollection(addressChannel, done, cnt)
}

func (a *AddressImportService) Index(houseRepos repository.HouseRepositoryInterface, isFull bool, start time.Time) {
	err := a.addressRepo.Index(houseRepos, isFull, start)
	if err != nil {
		a.logger.Error(err.Error())
	}
}

func (a *AddressImportService) ParseElement(decoder *xml.Decoder, element *xml.StartElement) (interface{}, error) {
	result := entity.AddressObject{}
	if element.Name.Local == "Object" {
		err := decoder.DecodeElement(&result, element)
		if err != nil {
			return nil, err
		}

		return result, nil
	}

	return nil, errors.New("object is not an address")
}
