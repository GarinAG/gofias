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

func (a *AddressImportService) Import(
	filePath string,
	wg *sync.WaitGroup,
	isFull bool,
	size int,
	insertCollection func(repo repository.InsertUpdateInterface, collection []interface{}, node interface{}, isFull bool, size int) []interface{},
) {

	defer wg.Done()
	start := time.Now()
	addressChannel := make(chan interface{})
	done := make(chan bool)
	//defer close(addressChannel)
	go util.ParseFile(filePath, addressChannel, done, a.logger, a.ParseElement)
	count := 0
	var collection []interface{}

Loop:
	for {
		select {
		case node := <-addressChannel:
			collection = insertCollection(a.addressRepo, collection, node, isFull, size)
			count++
		case <-done:
			break Loop
		}
	}
	if len(collection) > 0 {
		collection = insertCollection(a.addressRepo, collection, nil, isFull, size)
	}
	finish := time.Now()

	a.logger.Info("Number of addresses added: ", count)
	a.logger.Info("Time to import addresses: ", finish.Sub(start))
}

func (a *AddressImportService) Flush(wg *sync.WaitGroup, fool bool, params ...interface{}) {
	defer wg.Done()
	err := a.addressRepo.Flush(fool, params)
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
