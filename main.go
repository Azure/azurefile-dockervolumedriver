package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/go-plugins-helpers/volume"
)

const (
	volumeDriverName = "azurefile"
	mountpoint       = "/var/run/docker/volumedriver/azurefile"
	metadataRoot     = "/etc/docker/plugins/azurefile/volumes"
)

func main() {
	cmd := cli.NewApp()
	cmd.Name = "azurefile-dockervolumedriver"
	cmd.Version = "0.1"
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
		mountpoint := c.String("mountpoint")
		metaDir := c.String("metadata")
		removeShares := c.Bool("remove-shares")
		bindAddr := c.String("bind")
		if accountName == "" || accountKey == "" {
			log.Fatal("azure storage account name and key must be provided.")
		}

		log.WithFields(log.Fields{
			"accountName":  accountName,
			"metadata":     metaDir,
			"mountpoint":   mountpoint,
			"removeShares": removeShares,
		}).Debugf("Starting server at %s", bindAddr)

		driver, err := New(accountName, accountKey, mountpoint, metaDir, removeShares)
		if err != nil {
			log.Fatal(err)
		}
		h := volume.NewHandler(driver)
		log.Fatal(h.ServeUnix("root", volumeDriverName))
	}
	cmd.Run(os.Args)
}
