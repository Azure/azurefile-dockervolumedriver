# Docker Volume Driver for Azure File Service

This is a Docker Volume Driver which uses [Azure Storage File Service][afs]
to mount file shares on the cloud to Docker containers as volumes. It uses network
file sharing ([SMB/CIFS protocols][smb]) capabilities of Azure File Service.

## Why?

- You can create Docker containers that can migrate from one host to another seamlessly.
- You can share volumes among multiple containers running on different hosts.

## Usage

#### Start volume driver daemon

* Make sure you have a Storage Account on Azure (using Azure CLI or Portal).
* The server process must be running on the host machine where Docker engine is installed on 
  at all times for volumes to work properly.
* “cifs-utils” package must be installed on the host system as Azure Files use SMB protocol.
  For Debian/Ubuntu, run the following command on your host:
```shell
$ sudo apt-get install -y cifs-utils
```

You can start the server like this:

```shell
$ sudo ./azurefile-dockervolumedriver \
  --account-name <AzureStorageAccount> \
  --account-key  <AzureStorageAccountKey> &
```

> **NOTE:** Storage account must be in the same region as virtual machine. Otherwise
> you will get an error like “Host is down”.

Ideally you would want to run it on top of an init system (such as supervisord, systemd,
runit) that would start it automatically and keep it running in case of reboots and crashes.

#### Create volumes and containers

Starting from Docker 1.9+ you can create volumes and containers as follows:

```shell
$ docker volume create --name my_volume -d azurefile -o share=myshare
$ docker run -i -t -v my_volume:/data busybox
```

or simply:

```shell
$ docker run -it -v $(docker volume create -d azurefile -o share=myshare):/data busybox
```

This will create an Azure File Share named `myshare` (if it does not exist)
and start a Docker container in which you can use `/data` directory to directly
read/write from cloud file share location using SMB protocol.

## Command-line interface

```
$ ./azurefile-dockervolumedriver -h

NAME:
   azurefile-dockervolumedriver - Docker Volume Driver for Azure File Service

USAGE:
   azurefile-dockervolumedriver [options]

VERSION:
   0.1

OPTIONS:
   --account-name                                               Azure storage account name [$AZURE_STORAGE_ACCOUNT]
   --account-key                                                Azure storage account key [$AZURE_STORAGE_ACCOUNT_KEY]
   --remove-shares                                              remove associated Azure File Share when volume is removed
   --debug                                                      Enable verbose logging [$DEBUG]
   --mountpoint "/var/run/docker/volumedriver/azurefile"        Host path where volumes are mounted at
   --metadata "/etc/docker/plugins/azurefile/volumes"           Path where volume metadata are stored
   --bind ":8080"                                               Network addr to listen for requests
   --help, -h                                                   show help
   --version, -v
```

## Building

In order to compile this program, you need to have Go 1.5:

```sh
$ git clone https://github.com/ahmetalpbalkan/azurefile-dockervolumedriver src/azurefile
$ export GOPATH=`pwd`
$ export GO15VENDOREXPERIMENT=1
$ cd src/azurefile
$ go build -o azurefile
$ ./azurefile -h
```

## License

```
Copyright 2015 Ahmet Alp Balkan

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

[afs]: http://blogs.msdn.com/b/windowsazurestorage/archive/2014/05/12/introducing-microsoft-azure-file-service.aspx
[smb]: https://msdn.microsoft.com/en-us/library/windows/desktop/aa365233(v=vs.85).aspx
