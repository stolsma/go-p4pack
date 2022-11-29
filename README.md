<!--
/*
 * SPDX-FileCopyrightText: Copyright 2021-present Sander Tolsma. All rights reserved
 * SPDX-License-Identifier: Apache-2.0
 */
- -->

# Go-P4Pack: Generic packages & examples for Go & P4 based networking apps
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![License: CC BY-NC 4.0](https://img.shields.io/badge/License-CC_BY--NC_4.0-lightgrey.svg)](https://creativecommons.org/licenses/by-nc/4.0/)
[![Coverage Status](https://coveralls.io/repos/github/stolsma/go-p4pack/badge.svg?branch=main)](https://coveralls.io/github/stolsma/go-p4pack?branch=main)
[![Go-P4Pack Lint/Build/Test](https://github.com/stolsma/go-p4pack/actions/workflows/go-build-lint-test.yml/badge.svg)](https://github.com/stolsma/go-p4pack/actions/workflows/go-build-lint-test.yml)

Always wanted to write performant P4 based networking application in Go but don't know where to start? Then this is the place to get to. This repository contains several ready to use packages written in Go along with several example applications using those packages.
One of the larger (currently not ready) example applications is a p4Runtime/gNMI/gNOI API capable, Golang + DPDK SWX based, P4 programmable virtual soft switch. But also a gNMI CLI application and a bare bones DPDK SWX based dataplane switch (cmd/dpdkinfra) is included. Read the documentation below and start experimenting!

# Installation, build & run docker container

## Introduction

Prerequisites:
- User needs sudo execution permission


## Download repository and build container yourself

``` bash
git clone https://github.com/stolsma/go-p4pack.git
cd go-p4pack
```

Build go-p4pack docker image with all the example applications:

``` bash
./build/dpdk/build.sh 
./build/go-p4pack/build.sh 
```

If you did not do the local build, then pull the image from the GitHub Container Registry (ghcr.io):

``` bash
docker pull ghcr.io/stolsma/go-p4pack:latest
```

## Before startup: setup hugepages

Before you can run the example applications included in the go-p4pack docker image you first need to setup hugepages on your system.

``` bash
sudo mkdir /mnt/huge
sudo mount -t hugetlbfs nodev /mnt/huge
sudo sysctl -w vm.nr_hugepages=256
```

## Start go-p4pack docker container


Run go-p4pack docker image:

TODO: extend go-p4pack script for easy build & startup of the docker image

``` bash
./go-p4pack.sh
```

## Startup of the DPDK SWX Pipeline driver (cmd/dpdkinfra)

TODO: Describe how the different compiled example programs can be run!

The standard DPDK SWX example program can be run by:

``` bash
root@dec72f3353eb:/go-p4pack# ./dpdkinfra 
```

## Connect to the cmd/dpdkinfra driver integrated ssh terminal

From a second bash terminal to the docker host:

``` bash
ssh -p 2222 user@0.0.0.0
```

Or connect to the runing docker image with:

``` bash
docker exec -it go-p4pack /bin/bash
```

And create the demo netowk environment as described in the next paragraph.

## Test the Go DPDK SWX Pipeline driver (cmd/dpdkinfra)

Setup network namespaced test environment with two hosts:

``` bash
./examples/default/createenv.sh
```

This script will execute:

``` bash
sudo ip netns add host1
sudo ip link set sw1 netns host1
sudo ip netns exec host1 sudo ip link set lo up
sudo ip netns exec host1 sudo ip link set sw1 address 32:fb:fa:c6:67:01
sudo ip netns exec host1 sudo ip -4 addr add 192.168.222.1/24 dev sw1
sudo ip netns exec host1 sudo ip link set sw1 up
sudo ip netns exec host1 sudo arp -s 192.168.222.2 32:fb:fa:c6:67:02

sudo ip netns add host2
sudo ip link set sw2 netns host2
sudo ip netns exec host2 sudo ip link set lo up
sudo ip netns exec host2 sudo ip link set sw2 address 32:fb:fa:c6:67:02
sudo ip netns exec host2 sudo ip -4 addr add 192.168.222.2/24 dev sw2
sudo ip netns exec host2 sudo ip link set sw2 up
sudo ip netns exec host2 sudo arp -s 192.168.222.1 32:fb:fa:c6:67:01
```

Test connectivity

``` bash
sudo ip netns list
sudo ip netns exec host1 ping 192.168.222.2
sudo ip netns exec host2 ping 192.168.222.1
sudo ip netns exec host3 ping 192.168.222.1
sudo ip netns exec host4 ping 192.168.222.1
```

Do some performance testing

``` bash
# add -u to do UDP testing!
sudo ip netns exec host1 iperf -s
sudo ip netns exec host2 iperf -c 192.168.222.1 -P 10

sudo ip netns exec host3 iperf -s
sudo ip netns exec host4 iperf -c 192.168.222.3 -P 10

# with iperf3...
sudo ip netns exec host1 iperf3 -s
sudo ip netns exec host2 iperf3 -c 192.168.222.1 -P 10

# to dump packets send and received from interface
sudo ip netns exec host1 tcpdump -e -i sw1
```

Remove the test namespace environment

``` bash
sudo ip netns del host1
sudo ip netns del host2
```
