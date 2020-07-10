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

type HouseImportService struct {
	houseRepo repository.HouseRepositoryInterface
	logger    interfaces.LoggerInterface
}

func NewHouseService(houseRepo repository.HouseRepositoryInterface, logger interfaces.LoggerInterface) *HouseImportService {
	err := houseRepo.Init()
	if err != nil {
		logger.Panic(err.Error())
		os.Exit(1)
	}

	return &HouseImportService{
		houseRepo: houseRepo,
		logger:    logger,
	}
}

func (h *HouseImportService) GetRepo() repository.HouseRepositoryInterface {
	return h.houseRepo
}

func (h *HouseImportService) Import(
	filePath string,
	wg *sync.WaitGroup,
	isFull bool,
	size int,
	insertCollection func(repo repository.InsertUpdateInterface, collection []interface{}, node interface{}, isFull bool, size int) []interface{},
) {
	defer wg.Done()
	start := time.Now()
	houseChannel := make(chan interface{})
	done := make(chan bool)
	defer close(houseChannel)

	go util.ParseFile(filePath, houseChannel, done, h.logger, h.ParseElement)
	var collection []interface{}
	count := 0

Loop:
	for {
		select {
		case node := <-houseChannel:
			collection = insertCollection(h.houseRepo, collection, node, isFull, size)
			count++
		case <-done:
			break Loop
		}
	}
	if len(collection) > 0 {
		collection = insertCollection(h.houseRepo, collection, nil, isFull, size)
	}
	finish := time.Now()

	h.logger.Info("Number of homes added: ", count)
	h.logger.Info("Time to import houses: ", finish.Sub(start))
}

func (h *HouseImportService) ParseElement(decoder *xml.Decoder, element *xml.StartElement) (interface{}, error) {
	result := entity.HouseObject{}
	if element.Name.Local == "House" {
		err := decoder.DecodeElement(&result, element)
		if err != nil {
			return nil, err
		}

		return result, nil
	}

	return nil, errors.New("object is not a house")
}
