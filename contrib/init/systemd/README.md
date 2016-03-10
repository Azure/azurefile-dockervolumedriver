# CoreOS (systemd) Installation Instructions

The following instructions are for CoreOS, however with minor tweaks it can be adopted
on other distros using systemd as well.

Make sure you have docker-engine v1.10+ installed on your CoreOS image. Older images
do not support docker volumes as implemented by this volume driver.

Run the following commands:

1. `sudo -s`
2. `mkdir -p /opt/bin`
3. Download the binary at `/opt/bin/`: `wget -qO/opt/bin/azurefile-dockervolumedriver [url]`
4. Make it executable: `chmod +x /opt/bin/azurefile-dockervolumedriver`
5. Save `.default` file to `/etc/default/azurefile-dockervolumedriver`
6. Edit `/etc/default/azurefile-dockervolumedriver` with your credentials.
7. Save `.service` file to `/etc/systemd/system/azurefile-dockervolumedriver.service`
8. Run `systemctl daemon-reload`
9. `systemctl enable azurefile-dockervolumedriver` so that it starts automatically next time
10. Run `systemctl start azurefile-dockervolumedriver`
11. Check status via `systemctl status azurefile-dockervolumedriver`

Try by creating a volume:

    docker volume create -d azurefile --name myvol -o share=myvol

You can find the logs at `journalctl -fu azurefile-dockervolumedriver`.


