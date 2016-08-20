package main

import (
	"os"

	azure "github.com/Azure/azure-sdk-for-go/storage"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/go-plugins-helpers/volume"
)

const (
	volumeDriverName = "azurefile"
	mountpoint       = "/var/run/docker/volumedriver/azurefile"
	metadataRoot     = "/etc/docker/plugins/azurefile/volumes"
)

var (
	// GitSummary is provided at compile-time when built with govvv.
	// If the source tree corresponds to a tag, the tag name is used.
	// Otherwise, provides a string summarizing the state of git tree.
	GitSummary string
)

func main() {
	cmd := cli.NewApp()
	cmd.Name = "azurefile-dockervolumedriver"
	cmd.Version = GitSummary
	cmd.Usage = "Docker Volume Driver for Azure File Service"
	cli.AppHelpTemplate = usageTemplate

	cmd.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "account-name",
			Usage:  "Azure storage account name",
			EnvVar: "AZURE_STORAGE_ACCOUNT",
		},
		cli.StringFlag{
			Name:   "account-key",
			Usage:  "Azure storage account key",
			EnvVar: "AZURE_STORAGE_ACCOUNT_KEY",
		},
		cli.StringFlag{
			Name:   "storage-base",
			Usage:  "Base domain for Azure Storage endpoint",
			EnvVar: "AZURE_STORAGE_BASE",
			Value:  azure.DefaultBaseURL,
		},
		cli.BoolFlag{
			Name:  "remove-shares",
			Usage: "remove associated Azure File Share when volume is removed",
		},
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "Enable verbose logging",
			EnvVar: "DEBUG",
		},
		cli.StringFlag{
			Name:  "mountpoint",
			Usage: "Host path where volumes are mounted at",
			Value: mountpoint,
		},
		cli.StringFlag{
			Name:  "metadata",
			Usage: "Path where volume metadata are stored",
			Value: metadataRoot,
		},
	}
	cmd.Action = func(c *cli.Context) {
		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}

		accountName := c.String("account-name")
		accountKey := c.String("account-key")
		storageBase := c.String("storage-base")
		mountpoint := c.String("mountpoint")
		metaDir := c.String("metadata")
		removeShares := c.Bool("remove-shares")
		if accountName == "" || accountKey == "" {
			log.Fatal("azure storage account name and key must be provided.")
		}

		log.WithFields(log.Fields{
			"accountName":  accountName,
			"metadata":     metaDir,
			"mountpoint":   mountpoint,
			"removeShares": removeShares,
		}).Debug("Starting server.")

		driver, err := newVolumeDriver(accountName, accountKey, storageBase, mountpoint, metaDir, removeShares)
		if err != nil {
			log.Fatal(err)
		}
		h := volume.NewHandler(driver)
		log.Fatal(h.ServeUnix("docker", volumeDriverName))
	}
	cmd.Run(os.Args)
}
