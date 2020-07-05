package main

import (
	"flag"
	"log"
	"os"
	"runtime"
	"time"
)

func main() {
	flag.Parse()
	numcpu := runtime.NumCPU()
	runtime.GOMAXPROCS(numcpu)
	log.SetFlags(0)
	log.Printf("Snapshot storage place in: %s", *storage)
	log.Printf("Elasticsearch on %s", *host)
	log.Printf("Num CPU count: %d", numcpu)
	updateFias()
	log.Println("Fias import finished")
}

func updateFias(){
	doESConnection()
	getVersionInfo()
	if currentVersion.ID == lastDownloadVersion.VersionId && !*force {
		log.Println("Last version is uploaded")
		os.Exit(0)
	}
	if *force || !indexExists(addressIndexName) || currentVersion.ID == "" {
		startFullImport()
	} else {
		startDeltaImport()
	}
	createFullSnapshot()
	clearTmpDir()
}

func startFullImport()  {
	log.Println("")
	log.Printf("Start import full version: %s %s", lastDownloadVersion.VersionId, lastDownloadVersion.TextVersion)
	createTmpDir()
	importFull()
	updateInfo(lastDownloadVersion)
}

func startDeltaImport()  {
	getDownloadVersionList()
	if len(downloadVersionList) == 0 {
		log.Fatal("Versions not found in service")
	}
	var needVersionList []DownloadFileInfo
	for _, version := range downloadVersionList{
		if version.VersionId == currentVersion.ID {
			break
		}
		needVersionList = append(needVersionList, version)
	}

	for i := len(needVersionList) - 1; i >= 0; i-- {
		updateDelta(needVersionList[i])
	}
}

func updateDelta(version DownloadFileInfo)  {
	versionDateSlice := version.TextVersion[len(version.TextVersion) - 10: len(version.TextVersion)]
	versionTime, _ := time.Parse("02.01.2006", versionDateSlice)
	versionDate = versionTime.Format("2006-01-02") + dateTimeZone

	log.Println("")
	log.Printf("Start update index to version: %s %s", version.VersionId, version.TextVersion)
	createTmpDir()
	update(version.FiasDeltaXmlUrl)
	updateInfo(version)
}

func importFull()  {
	downloadFull()
	if !*skipHouses {
		housesFullImport()
	}
	addressesFullImport()
}

func update(fileUrl string)  {
	isUpdate = true
	downloadUpdate(fileUrl)
	if !*skipHouses {
		housesDeltaImport()
	}
	addressesDeltaImport()
}