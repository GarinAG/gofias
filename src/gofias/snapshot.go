package main

import (
	"context"
	"fmt"
	"strings"
)

const (
	repositoryBody = `
    {
      "type": "fs",
      "settings": {
        "compress": "true",
        "location": "%location%"
      }
    }
`
	snapshotBody = `
    {
      "indices": ["%address_index%", "%house_index%"],
      "ignore_unavailable": "true",
      "include_global_state": "false",
      "metadata": {
        "taken_by": "fias",
        "taken_because": "backup before update"
      }
    }
`
)

func registerRepository() {
	_, err := elasticClient.SnapshotCreateRepository(GetPrefixIndexName(repositoryName)).
		BodyString(strings.ReplaceAll(repositoryBody, "%location%", *storage)).
		Do(context.Background())

	if err != nil {
		logFatal(err)
	}
}

func restoreFromSnapshot(force bool) {
	ctx := context.Background()
	reposName := GetPrefixIndexName(repositoryName)
	snapName := GetPrefixIndexName(snapshotName)
	addressPrefixIndexName := GetPrefixIndexName(addressIndexName)
	housePrefixIndexName := GetPrefixIndexName(houseIndexName)

	snapshot, _ := elasticClient.SnapshotGet(reposName).
		Snapshot(snapName).
		Do(ctx)

	if len(snapshot.Snapshots) > 0 {
		if force {
			_, _ = elasticClient.DeleteIndex(addressPrefixIndexName, housePrefixIndexName).
				Do(ctx)
		}

		_, err := elasticClient.SnapshotRestore(reposName, snapName).
			BodyString(fmt.Sprintf(`{"indices": ["%s", "%s"]}`, addressPrefixIndexName, housePrefixIndexName)).
			Do(ctx)

		if err != nil {
			logFatal(err)
		}
	}
}

func createFullSnapshot() {
	logPrintln("Create full snapshot")
	ctx := context.Background()
	reposName := GetPrefixIndexName(repositoryName)
	snapName := GetPrefixIndexName(snapshotName)
	addressPrefixIndexName := GetPrefixIndexName(addressIndexName)
	housePrefixIndexName := GetPrefixIndexName(houseIndexName)

	repository, _ := elasticClient.SnapshotGetRepository(reposName).Do(ctx)
	if repository == nil {
		registerRepository()
	} else {
		snapshot, _ := elasticClient.SnapshotGet(reposName).Snapshot(snapName).Do(ctx)
		if snapshot != nil {
			_, err := elasticClient.SnapshotDelete(reposName, snapName).Do(ctx)
			if err != nil {
				logPrintln(err)
			}
		}
	}

	snapBody := strings.ReplaceAll(snapshotBody, "%address_index%", addressPrefixIndexName)
	snapBody = strings.ReplaceAll(snapBody, "%house_index%", housePrefixIndexName)

	_, err := elasticClient.SnapshotCreate(reposName, snapName).BodyString(snapBody).Do(ctx)
	if err != nil {
		logPrintln(err)
	}
}
