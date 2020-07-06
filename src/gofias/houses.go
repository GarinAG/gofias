package main

import (
	"path/filepath"
)

func housesFullImport() {
	unzipHouses()
	DropIndex(houseIndexName)
	CreateIndex(houseIndexName, houseIndexSettings)
	CreatePreprocessor(housesPipeline, houseDropPipeline)
	searchAndImportHouses()
	RefreshIndexes()
}

func housesDeltaImport() {
	unzipHousesDelta()
	searchAndImportHouses()
	RefreshIndexes()
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
		logPrintln(err)
	}
	var total uint64 = 0
	if len(matches) != 0 {
		for _, v := range matches {
			total += importHouse(v)
		}
	}

	recUpdateHouses = total
}
