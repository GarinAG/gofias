package main

import (
	"flag"
	"os"
	"runtime"
	"time"
)

func main() {
	flag.Parse()
	setNumCpu()
	setLogPath()
	logPrintf("Snapshot storage place in: %s", *storage)
	logPrintf("Elasticsearch on %s", *host)
	logPrintf("Num CPU count: %d", *numCpu)
	updateFias()
	logPrintln("Fias import finished")
}

func setNumCpu() {
	if *numCpu == 0 {
		*numCpu = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(*numCpu)
}

func updateFias() {
	DoESConnection()
	getVersionInfo()
	if *forceIndex {
		if IndexExists(addressIndexName) {
			CreateAddressIndex()
		} else {
			logPrintln("Address index bot found")
		}
		os.Exit(0)
	}
	if currentVersion.ID == lastDownloadVersion.VersionId && !*force {
		logPrintln("Last version is uploaded")
		os.Exit(0)
	}
	if *force || !IndexExists(addressIndexName) || currentVersion.ID == "" {
		startFullImport()
	} else {
		startDeltaImport()
	}
	if !*skipSnapshot {
		createFullSnapshot()
	}
	clearTmpDir()
	logPrintln("Import Finished")
}

func startFullImport() {
	logPrintf("Start import full version: %s %s", lastDownloadVersion.VersionId, lastDownloadVersion.TextVersion)
	createTmpDir()
	importFull()
	updateInfo(lastDownloadVersion)
}

func startDeltaImport() {
	getDownloadVersionList()
	if len(downloadVersionList) == 0 {
		logFatal("Versions not found in service")
	}
	var needVersionList []DownloadFileInfo
	for _, version := range downloadVersionList {
		if version.VersionId == currentVersion.ID {
			break
		}
		needVersionList = append(needVersionList, version)
	}

	for i := len(needVersionList) - 1; i >= 0; i-- {
		updateDelta(needVersionList[i])
	}
}

func updateDelta(version DownloadFileInfo) {
	versionDateSlice := version.TextVersion[len(version.TextVersion)-10 : len(version.TextVersion)]
	versionTime, _ := time.Parse("02.01.2006", versionDateSlice)
	versionDate = versionTime.Format("2006-01-02") + dateTimeZone

	logPrintf("Start update index to version: %s %s", version.VersionId, version.TextVersion)
	createTmpDir()
	update(version.FiasDeltaXmlUrl)
	updateInfo(version)
	*skipClear = false
}

func importFull() {
	downloadFull()
	if !*skipHouses {
		housesFullImport()
	}
	addressesFullImport()
}

func update(fileUrl string) {
	isUpdate = true
	downloadUpdate(fileUrl)
	if !*skipHouses {
		housesDeltaImport()
	}
	addressesDeltaImport()
}
