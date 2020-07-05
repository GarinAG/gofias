package main

import (
	"log"
	"path/filepath"
)

func housesFullImport() {
	unzipHouses()
	dropIndex(houseIndexName)
	createIndex(houseIndexName, houseIndexSettings)
	createPreprocessor(housesPipeline, houseDropPipeline)
	searchAndImportHouses()
	refreshIndexes()
}

func housesDeltaImport() {
	unzipHousesDelta()
	searchAndImportHouses()
	refreshIndexes()
}

func unzipHouses() {
	err := Unzip(tmpDirPath+fiasXml, tmpDirPath, housesFilePart)
	if err != nil {
		panic(err)
	}
}

func unzipHousesDelta() {
	err := Unzip(tmpDirPath+fiasDeltaXml, tmpDirPath, housesFilePart)
	if err != nil {
		panic(err)
	}
}

func searchAndImportHouses() {
	matches, err := filepath.Glob(tmpDirPath + housesFilePart)
	if err != nil {
		log.Println(err)
	}
	var total uint64 = 0
	if len(matches) != 0 {
		for _, v := range matches {
			total += importHouse(v)
		}
	}

	recUpdateHouses = total
}
