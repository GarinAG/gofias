package main

import (
	"path/filepath"
)

func addressesFullImport() {
	unzipAddresses()
	DropIndex(addressIndexName)
	CreateIndex(addressIndexName, addrIndexSettings)
	CreatePreprocessor(addrPipeline, addrDropPipeline)
	searchAndImportAddresses()
	CreateAddressIndex()
	RefreshIndexes()
}

func addressesDeltaImport() {
	unzipAddressesDelta()
	CreatePreprocessor(addrPipeline, addrDropPipeline)
	searchAndImportAddresses()
	CreateAddressIndex()
	RefreshIndexes()
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
		logPrintln(err)
	}
	var total uint64 = 0
	if len(matches) != 0 {
		for _, v := range matches {
			total += importAddress(v)
		}
	}

	recUpdateAddress = total
}
