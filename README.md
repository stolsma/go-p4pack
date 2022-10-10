<!--
/*
 * Copyright 2021 - Sander Tolsma. All rights reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */
- -->

# Go-P4Pack: Generic packages & examples for Go & P4 based networking apps
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![License: CC BY-NC 4.0](https://img.shields.io/badge/License-CC_BY--NC_4.0-lightgrey.svg)](https://creativecommons.org/licenses/by-nc/4.0/)
[![Coverage Status](https://coveralls.io/repos/github/stolsma/go-p4pack/badge.svg?branch=main)](https://coveralls.io/github/stolsma/go-p4pack?branch=main)
[![Go-P4Pack Lint/Build/Test](https://github.com/stolsma/go-p4pack/actions/workflows/go-build-lint-test.yml/badge.svg)](https://github.com/stolsma/go-p4pack/actions/workflows/go-build-lint-test.yml)

Always wanted to write performant P4 based networking application in Go but don't know where to start? Then this is the place to get to. This repository contains several ready to use packages written in Go along with several example applications using those packages.
One of the larger (currently not ready) example applications is a p4Runtime/gNMI/gNOI API capable, Golang + DPDK SWX based, P4 programmable virtual soft switch. But also a gNMI CLI application and a bare bones DPDK SWX based dataplane switch is included. Read the documentation below and start experimenting!

# Installation, build & run docker container

## Download repository

``` bash
git clone https://github.com/stolsma/go-p4pack.git
cd go-p4pack
```

Build go-p4pack docker image with all the example applications:

``` bash
./build/go-p4pack/build.sh 
```

## Setup hugepages

``` bash
sudo mkdir /mnt/huge
sudo mount -t hugetlbfs nodev /mnt/huge
sudo sysctl -w vm.nr_hugepages=256
```

## start docker container

Run go-p4pack docker image:

``` bash
./go-p4pack run
```

## Startup the DPDK SWX Pipeline driver

TODO: Describe how the different compiled example programs can be run!

The standard DPDK SWX example program can be run by:

``` bash
sudo ./dpdk-pipeline -c 0x3 -- -s ./examples/ipdk-simple_l3/simple_l3.cli
# sudo ./dpdk-pipeline -c 0x3 --log-level='.*,8' -- -s ./examples/ipdk-simple_l3/simple_l3.cli
```

## Connect to the driver ssh terminal

from a second bash terminal

``` bash
ssh -p 2222 user@0.0.0.0
```

## Test the Go DPDK SWX Pipeline driver example

Setup network namespaced test environment with two hosts:

``` bash
sudo ip netns add host0
sudo ip link set sw0 netns host0
sudo ip netns exec host0 sudo ip link set lo up
sudo ip netns exec host0 sudo ip link set sw0 address 32:fb:fa:c6:67:1f
sudo ip netns exec host0 sudo ip -4 addr add 192.168.222.1/24 dev sw0
sudo ip netns exec host0 sudo ip link set sw0 up
sudo ip netns exec host0 sudo arp -s 192.168.222.2 e6:48:78:26:12:52

sudo ip netns add host1
sudo ip link set sw1 netns host1
sudo ip netns exec host1 sudo ip link set lo up
sudo ip netns exec host1 sudo ip link set sw1 address e6:48:78:26:12:52
sudo ip netns exec host1 sudo ip -4 addr add 192.168.222.2/24 dev sw1
sudo ip netns exec host1 sudo ip link set sw1 up
sudo ip netns exec host1 sudo arp -s 192.168.222.1 32:fb:fa:c6:67:1f
```

Test connectivity

``` bash
sudo ip netns list
sudo ip netns exec host0 ping 192.168.222.2
sudo ip netns exec host1 ping 192.168.222.1
```

Do some performance testing

``` bash
# add -u to do UDP testing!
sudo ip netns exec host0 iperf -s
sudo ip netns exec host1 iperf -c 192.168.222.1 -P 10
```

Remove the test namespace environment

``` bash
sudo ip netns del host0
sudo ip netns del host1
```

# Development Scratch Space

TODO Remove or use in other parts of this README!

``` bash
pipeline PIPELINE0 stats
Input ports:
        Port 0: packets 0 bytes 0 empty 0
        Port 1: packets 0 bytes 0 empty 0
        Port 2: packets 0 bytes 0 empty 0
        Port 3: packets 0 bytes 0 empty 0

Output ports:
        Port 0: packets 0 bytes 0
        Port 1: packets 0 bytes 0
        Port 2: packets 0 bytes 0
        Port 3: packets 0 bytes 0
        DROP: packets 0 bytes 0

Tables:
        Table ipv4_host:
                Hit (packets): 0
                Miss (packets): 0
                Action NoAction (packets): 0
                Action send (packets): 0
                Action drop_1 (packets): 0

Learner tables:
pipeline>


pipeline PIPELINE0 table ipv4_host add ./examples/ipdk-simple_l3/l3_table.txt

pipeline PIPELINE0 table ipv4_host show <filename>



sudo ./dpdk-pipeline -c 0x3 -- -s ./examples/vxlan/vxlan_pcap.cli

sudo ./dpdk-pipeline -c 0x3 --vdev=net_tap0,iface=sw0 --vdev=net_tap1,iface=sw1  --vdev=net_tap2,iface=sw2  --vdev=net_tap3,iface=sw3 -- -s ./examples/vxlan/vxlan.cli

telnet 0.0.0.0 8086


docker run -it --device=/dev/vfio/48 --device=/dev/vfio/49 --device=/dev/vfio/vfio --ulimit memlock=-1 -v /dev/hugepages:/dev/hugepages krsna1729/dpdk-l2fwd-bin



docker run --rm -u 1000:1000 -v ${PWD}:/p4ccode -w /p4ccode stolsma/p4c-all:latest /bin/bash -c "p4c-dpdk -o vxlan.spec --p4runtime-files p4info.proto.txt vxlan.p4"

sudo ../../dpdk-pipeline -c 0x3 -- -s ./vxlan_pcap.cli

sudo ../../dpdk-pipeline -c 0x3 --vdev=net_tap0,iface=sw0 --vdev=net_tap1,iface=sw1  --vdev=net_tap2,iface=sw2  --vdev=net_tap3,iface=sw3 -- -s ./vxlan.cli

```

# DPDK-Pipeline CLI Commands

``` bash
pipeline> help
Type 'help <command>' for command details.

List of commands:
        mempool
        link
        tap
        pipeline create
        pipeline port in
        pipeline port out
        pipeline build
        pipeline table add
        pipeline table delete
        pipeline table default
        pipeline table show
        pipeline selector group add
        pipeline selector group delete
        pipeline selector group member add
        pipeline selector group member delete
        pipeline selector show
        pipeline learner default
        pipeline commit
        pipeline abort
        pipeline regrd
        pipeline regwr
        pipeline meter profile add
        pipeline meter profile delete
        pipeline meter reset
        pipeline meter set
        pipeline meter stats
        pipeline stats
        thread pipeline enable
        thread pipeline disable

mempool <mempool_name> buffer <buffer_size> pool <pool_size> cache <cache_size> cpu <cpu_id>
link <link_name> {dev <device_name> | port <port_id>} rxq <n_queues> <queue_size> <mempool_name> txq <n_queues> <queue_size> promiscuous {on | off} [rss <qid_0> ... <qid_n>]
link show [<link_name>]
ring <ring_name> size <size> numa <numa_node>
tap <tap_name>
pipeline <pipeline_name> create <numa_node>
pipeline <pipeline_name> port in <port_id> {link <link_name> rxq <queue_id> bsz <burst_size> | ring <ring_name> bsz <burst_size> | source <mempool_name> <file_name> loop <n_loops> | tap <tap_name> mempool <mempool_name> mtu <mtu> bsz <burst_size>
pipeline <pipeline_name> port out <port_id> {link <link_name> txq <txq_id> bsz <burst_size> | ring <ring_name> bsz <burst_size> | sink <file_name> | none | tap <tap_name> bsz <burst_size>
pipeline <pipeline_name> build <spec_file>
pipeline <pipeline_name> table <table_name> add <file_name>
pipeline <pipeline_name> table <table_name> delete <file_name>
pipeline <pipeline_name> table <table_name> default <file_name>
pipeline <pipeline_name> table <table_name> show [filename]
pipeline <pipeline_name> selector <selector_name> group add
pipeline <pipeline_name> selector <selector_name> group delete <group_id>
pipeline <pipeline_name> selector <selector_name> group member add <file_name>
pipeline <pipeline_name> selector <selector_name> group member delete <file_name>
pipeline <pipeline_name> selector <selector_name> show [filename]
pipeline <pipeline_name> learner <learner_name> default <file_name>
pipeline <pipeline_name> commit
pipeline <pipeline_name> abort
pipeline <pipeline_name> regrd <register_array_name> <index
pipeline <pipeline_name> regwr <register_array_name> <index> <value>
pipeline <pipeline_name> meter profile <profile_name> add cir <cir> pir <pir> cbs <cbs> pbs <pbs>
pipeline <pipeline_name> meter profile <profile_name> delete
pipeline <pipeline_name> meter <meter_array_name> from <index0> to <index1> reset
pipeline <pipeline_name> meter <meter_array_name> from <index0> to <index1> set profile <profile_name>
pipeline <pipeline_name> meter <meter_array_name> from <index0> to <index1> stats
pipeline <pipeline_name> stats
thread <thread_id> pipeline <pipeline_name> enable
thread <thread_id> pipeline <pipeline_name> disable
```
