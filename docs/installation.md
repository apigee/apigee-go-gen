# Installation 
<!--
  Copyright 2024 Google LLC

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
-->

You can install the `apigee-go-gen` binary in several ways. 

### Manual download

The most straightforward way to install the tool is to download a release tar ball from the
available GitHub [releases](https://github.com/micovery/apigee-go-gen/releases).

Once you download the tarball, extract it and move the `apigee-go-gen` binary to somewhere in your `$PATH`

### Automated install

For your convenience, there is an [install](https://github.com/micovery/apigee-go-gen/blob/main/install) script available.

This script downloads and installs the `apigee-go-gen` tool automatically for you.

The script takes **version**, and install **directory** as optional parameters

> - [x] If **version** is omitted, it will download the latest tagged release.
- [x] If install **directory** is omitted, it will install to `/usr/local/bin`
- [x] If install **directory** is not writable, it will prompt you for sudo password.


Below are a few examples of how to execute the `install` script

e.g.

Install latest **version** into `/usr/local/bin` **directory**
```shell
curl -s https://raw.githubusercontent.com/micovery/apigee-go-gen/main/install | sh
```


Install specific **version** into `/usr/local/bin` **directory**
```shell
curl -s https://raw.githubusercontent.com/micovery/apigee-go-gen/main/install | sh -s v0.1.13
```

Install **latest** version into `~/.local/bin` **directory**
```shell
curl -s https://raw.githubusercontent.com/micovery/apigee-go-gen/main/install | sh -s latest ~/.local/bin
```

Install specific **version** into `~/.local/bin` **directory**
```shell
curl -s https://raw.githubusercontent.com/micovery/apigee-go-gen/main/install | sh -s v0.1.13 ~/.local/bin
```



### From source

If you already have [Go](https://go.dev/doc/install) installed in your machine, run the following command:

```shell
go install github.com/micovery/apigee-go-gen/cmd/...@latest
```

This will download the source, build it (in your machine) and install the `apigee-go-gen` binary into your `$GOPATH/bin` directory.

You can change the `@latest` tag for any other version that has been tagged. (e.g. `@v0.1.13`)

!!! Note
    The Go tool (and compiler) is only necessary to build the tools in this repo.
    Once built, you can copy the tool binaries and use them in any other
    machine of the same architecture and operating system (without needing Go).
