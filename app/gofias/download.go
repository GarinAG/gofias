package main

import (
	"os"
	"os/user"
)

var (
	tmpDirPath string
)

func createTmpDir() {
	usr, err := user.Current()
	if err != nil {
		logFatal(err)
	}
	dir := usr.HomeDir + *tmp

	logPrintf("Tmp dir place in: %s", dir)
	if !*skipClear {
		clearTmpDir()
	}
	if _, err := os.Stat(dir); err != nil {
		logPrintln("Create new Tmp dir")
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			logFatal(err)
		}
	}

	tmpDirPath = dir
}

func clearTmpDir() {
	usr, err := user.Current()
	if err != nil {
		logFatal(err)
	}
	dir := usr.HomeDir + *tmp
	if _, err := os.Stat(dir); err == nil {
		logPrintln("Clear Tmp dir")
		err = os.RemoveAll(dir)
		if err != nil {
			logFatal(err)
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
			logFatal(err)
		}
	}
}

func downloadUpdate(fileUrl string) {
	fileName := tmpDirPath + fiasDeltaXml
	if _, err := os.Stat(fileName); err != nil {
		err := DownloadFile(fileName, fileUrl)
		if err != nil {
			logFatal(err)
		}
	}
}
