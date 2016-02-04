# Docker Volume Driver for Azure File Service

This is a Docker Volume Driver which uses [Azure Storage File Service][afs]
to mount file shares on the cloud to Docker containers as volumes. It uses network
file sharing ([SMB/CIFS protocols][smb]) capabilities of Azure File Service.

[![Build Status](https://travis-ci.org/Azure/azurefile-dockervolumedriver.svg?branch=master)](https://travis-ci.org/Azure/azurefile-dockervolumedriver)

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

Please refer to “Building” section below on how to compile.

Once you have the volume driver compiled you can start it as a server:

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

## Demo

![](http://cl.ly/image/2z1z1y030u3B/Image%202015-10-06%20at%203.18.39%20PM.gif)

## Changelog

```

# 0.1 (2016-02-08)
- Initial release.

```

## Building

If you need to use this project, please consider downloading it from “Releases”
link above. The following instructions are for compiling the project from source.

In order to compile this program, you need to have Go 1.5:

```sh
$ git clone https://github.com/Azure/azurefile-dockervolumedriver src/azurefile
$ export GOPATH=`pwd`
$ export GO15VENDOREXPERIMENT=1
$ cd src/azurefile
$ go build -o azurefile
$ ./azurefile -h
```

### Installing on Ubuntu 14.04 or lower

Please [check out the docs](contrib/init/upstart/README.md) to configure
the service for Ubuntu distributions with upstart.

## Author

* [Ahmet Alp Balkan](https://github.com/ahmetalpbalkan)

## License

```
Copyright 2015 Microsoft Corporation

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
