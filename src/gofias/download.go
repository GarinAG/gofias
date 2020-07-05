package main

import (
	"log"
	"os"
	"os/user"
)

var (
	tmpDirPath string
)

func createTmpDir() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	dir := usr.HomeDir + *tmp

	log.Printf("Tmp dir place in: %s", dir)
	clearTmpDir()
	if _, err := os.Stat(dir); err != nil {
		log.Println("Create new Tmp dir")
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	tmpDirPath = dir
}

func clearTmpDir() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	dir := usr.HomeDir + *tmp
	if _, err := os.Stat(dir); err == nil {
		log.Println("Clear Tmp dir")
		err = os.RemoveAll(dir)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getVersionInfo() {
	getLastVersion()
	getLastDownloadVersion()
}

func downloadFull() {
	fileName := tmpDirPath + fiasXml
	if _, err := os.Stat(fileName); err != nil {
		err := DownloadFile(fileName, urlFullPath)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func downloadUpdate(fileUrl string) {
	fileName := tmpDirPath + fiasDeltaXml
	if _, err := os.Stat(fileName); err != nil {
		err := DownloadFile(fileName, fileUrl)
		if err != nil {
			log.Fatal(err)
		}
	}
}
