package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"
	"github.com/ahmetalpbalkan/dkvolume"
)

const (
	volumeDriverName = "azurefile"
	mountpoint       = "/var/run/docker/volumedriver/azurefile"
	metadataRoot     = "/etc/docker/plugins/azurefile/volumes/"
)

var (
	flDebug           = flag.Bool("debug", false, "enable verbose logging")
	flMountpoint      = flag.String("mountpoint", mountpoint, "volume mount point path")
	flMetaDir         = flag.String("metadata-path", metadataRoot, "volume mount point path")
	flAzureAccount    = flag.String("account-name", "", "Azure storage account name, used as default if not specified per volume")
	flAzureAccountKey = flag.String("account-key", "", "Azure storage account key, used as default if not specified per volume")
	flRemoveShares    = flag.Bool("remove-shares", false, "remove associated azure file share when volume is removed")
)

func main() {
	flag.Parse()
	if *flDebug {
		log.SetLevel(log.DebugLevel)
	}

	if *flAzureAccount == "" || *flAzureAccountKey == "" {
		log.Fatal("azure storage account name and key must be provided.")
	}

	driver, err := New(*flAzureAccount, *flAzureAccountKey, *flMountpoint, *flMetaDir, *flRemoveShares)
	if err != nil {
		log.Fatal(err)
	}
	h := dkvolume.NewHandler(driver)
	log.WithFields(log.Fields{
		"accountName":  *flAzureAccount,
		"metaDir":      *flMetaDir,
		"mountpoint":   *flMountpoint,
		"removeShares": *flRemoveShares,
	}).Debug("Starting server.")

	log.Fatal(h.ServeTCP("azurefile", ":8080"))
}
