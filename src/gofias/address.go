package main

import (
	"log"
	"path/filepath"
)

func addressesFullImport() {
	unzipAddresses()
	dropIndex(addressIndexName)
	createIndex(addressIndexName, addrIndexSettings)
	createPreprocessor(addrPipeline, addrDropPipeline)
	searchAndImportAddresses()
	createAddressIndex()
	refreshIndexes()
}

func addressesDeltaImport() {
	unzipAddressesDelta()
	createPreprocessor(addrPipeline, addrDropPipeline)
	searchAndImportAddresses()
	createAddressIndex()
	refreshIndexes()
}

func unzipAddresses() {
	err := Unzip(tmpDirPath+fiasXml, tmpDirPath, addrFilePart)
	if err != nil {
		panic(err)
	}
}

func unzipAddressesDelta() {
	err := Unzip(tmpDirPath+fiasDeltaXml, tmpDirPath, addrFilePart)
	if err != nil {
		panic(err)
	}
}

func searchAndImportAddresses() {
	matches, err := filepath.Glob(tmpDirPath + addrFilePart)
	if err != nil {
		log.Println(err)
	}
	var total uint64 = 0
	if len(matches) != 0 {
		for _, v := range matches {
			total += importAddress(v)
		}
	}

	recUpdateAddress = total
}
