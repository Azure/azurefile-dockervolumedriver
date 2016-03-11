# Ubuntu (systemd) installation instructions

> NOTE: These instructions are valid for Ubuntu versions using `systemd` init
> system.

Run the following commands:

0. `sudo -s`
0. Download the binary at `/opt/bin/`: `wget -qO/usr/bin/azurefile-dockervolumedriver [url]`
0. Make it executable: `chmod +x /usr/bin/azurefile-dockervolumedriver`
0. Save the `.default` file to `/etc/default/azurefile-dockervolumedriver`
0. Edit `/etc/default/azurefile-dockervolumedriver` with your credentials.
0. Save the `.service` file to `/etc/systemd/system/azurefile-dockervolumedriver.service`
0. Run `systemctl daemon-reload`
0. Run `systemctl enable azurefile-dockervolumedriver`
0. Run `systemctl start azurefile-dockervolumedriver`
0. Check status via `systemctl status azurefile-dockervolumedriver`

Try by creating a volume and running a container with it:

    docker volume create -d azurefile --name myvol -o share=myvol
    docker run -i -t -v myvol:/data busybox
    # cd /data
    # touch a.txt

You can find the logs at `journalctl -fu azurefile-dockervolumedriver`.
