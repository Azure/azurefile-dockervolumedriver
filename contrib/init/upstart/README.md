# Ubuntu (upstart) installation instructions

These instructions are valid for Ubuntu versions using `upstart` init system.

### Step 1. Build or Download

Use instructions in the repository root to compile a binary for Linux or
go to “Releases” section to download a version.

Place the binary to `/usr/bin/azurefile-dockervolumedriver` and make it
executable with `chmod +x <path>`.

### Step 2. Copy init script

Copy the `.conf` file in this directory to `/etc/init/azurefile-dockervolumedriver.conf`
of the Ubuntu machine.

### Step 3. Copy the configuration file

Copy the `.default` file in this directory to `/etc/default/azurefile-dockervolumedriver`,
without trailing extension.

Open the file and edit the storage credentials to be used in the virtual machine.

### Step 4. Start the process

Once the files are copied, run these commands as sudo:

    initctl reload-configuration
    initctl start azurefile-dockervolumedriver

Now the volume driver plugin service should be started on the machine. Verify by running:

    initctl status azurefile-dockervolumedriver

and you should see an output saying “start/running” for the service.

From this point on every time the plugin service crashes or the system reboots, it should
be stated again by upstart.

### Step 5. Validate

Create a volume using docker CLI and create a container with this volume to see if you can
write to the Azure File Service share.

    docker volume create -d azurefile -o share=myshare --name=myvol
    docker run -i -t -v myvol:/data busybox
    (inside the container)
    # cd /data
    # touch file.txt

