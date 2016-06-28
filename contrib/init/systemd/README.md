# Ubuntu (systemd) installation instructions

> NOTE: These instructions are valid for Ubuntu versions using `systemd` init
> system.

## TL;DR
1. Get the latest [release](https://github.com/Azure/azurefile-dockervolumedriver/releases)
2. Put the binary into `/usr/bin/azurefile-dockervolumedriver`
3. Get the .default and .service files and deploy them
4. Reload systemd

## In-depth walkthrough

0. `sudo -s`
0. Use wget to get the `azurefile-dockervolumedriver.default` and `azurefile-dockervolumedriver.service` files from GitHub. These are in the `../contrib/init/systemd` directory.
0. Download the binary from the "Releases" tab of the repo to `/usr/bin/azurefile-dockervolumedriver`
    + Use wget to download to dir: `wget -qO/usr/bin/azurefile-dockervolumedriver https://github.com/Azure/azurefile-dockervolumedriver/releases/download/[VERSION]/azurefile-dockervolumedriver`
    + Make it executable `chmod +x /usr/bin/azurefile-dockervolumedriver`
0. Save the `.default` file to `/etc/default/azurefile-dockervolumedriver`
0. Edit `/etc/default/azurefile-dockervolumedriver` with your Azure Storage Account credentials.
0. Save the `.service` file to `/etc/systemd/system/azurefile-dockervolumedriver.service`
    + [Ubuntu 15.x only] Make the requisite directories if they don't exist: `mkdir -p /etc/systemd/system`
0. Run `systemctl daemon-reload`
0. Run `systemctl enable azurefile-dockervolumedriver`
0. Run `systemctl start azurefile-dockervolumedriver`
0. Check status via `systemctl status azurefile-dockervolumedriver`

To test, try to create a volume and running a container with it:

    docker volume create -d azurefile --name myvol -o share=myvol
    docker run -i -t -v myvol:/data busybox
    # cd /data
    # touch a.txt

You can find the logs at `journalctl -fu azurefile-dockervolumedriver`.
