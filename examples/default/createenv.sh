#!/usr/bin/env bash
#
# Copyright 2022 - Sander Tolsma. All rights reserved
# SPDX-License-Identifier: Apache-2.0

# let alias work in non interactive bash shells!
shopt -s expand_aliases

# Builds are run as root in containers, no need for sudo
[ "$(id -u)" != '0' ] || alias sudo=

# Remove leftovers before going on...
sudo ip netns del host1
sudo ip netns del host2
sudo ip netns del host3
sudo ip netns del host4

# Create host1
sudo ip netns add host1
sudo ip link set sw1 netns host1
sudo ip netns exec host1 sudo ip link set lo up
sudo ip netns exec host1 sudo ip link set sw1 address 32:fb:fa:c6:67:01
sudo ip netns exec host1 sudo ip -4 addr add 192.168.222.1/24 dev sw1
sudo ip netns exec host1 sudo ip link set sw1 up
sudo ip netns exec host1 sudo arp -s 192.168.222.2 32:fb:fa:c6:67:02
sudo ip netns exec host1 sudo arp -s 192.168.222.3 32:fb:fa:c6:67:03
sudo ip netns exec host1 sudo arp -s 192.168.222.4 32:fb:fa:c6:67:04

# Create host2
sudo ip netns add host2
sudo ip link set sw2 netns host2
sudo ip netns exec host2 sudo ip link set lo up
sudo ip netns exec host2 sudo ip link set sw2 address 32:fb:fa:c6:67:02
sudo ip netns exec host2 sudo ip -4 addr add 192.168.222.2/24 dev sw2
sudo ip netns exec host2 sudo ip link set sw2 up
sudo ip netns exec host2 sudo arp -s 192.168.222.1 32:fb:fa:c6:67:01
sudo ip netns exec host2 sudo arp -s 192.168.222.3 32:fb:fa:c6:67:03
sudo ip netns exec host2 sudo arp -s 192.168.222.4 32:fb:fa:c6:67:04

# Create host3
sudo ip netns add host3
sudo ip link set sw3 netns host3
sudo ip netns exec host3 sudo ip link set lo up
sudo ip netns exec host3 sudo ip link set sw3 address 32:fb:fa:c6:67:03
sudo ip netns exec host3 sudo ip -4 addr add 192.168.222.3/24 dev sw3
sudo ip netns exec host3 sudo ip link set sw3 up
sudo ip netns exec host3 sudo arp -s 192.168.222.1 32:fb:fa:c6:67:01
sudo ip netns exec host3 sudo arp -s 192.168.222.2 32:fb:fa:c6:67:02
sudo ip netns exec host3 sudo arp -s 192.168.222.4 32:fb:fa:c6:67:04

# Create host4
sudo ip netns add host4
sudo ip link set sw4 netns host4
sudo ip netns exec host4 sudo ip link set lo up
sudo ip netns exec host4 sudo ip link set sw4 address 32:fb:fa:c6:67:04
sudo ip netns exec host4 sudo ip -4 addr add 192.168.222.4/24 dev sw4
sudo ip netns exec host4 sudo ip link set sw4 up
sudo ip netns exec host4 sudo arp -s 192.168.222.1 32:fb:fa:c6:67:01
sudo ip netns exec host4 sudo arp -s 192.168.222.2 32:fb:fa:c6:67:02
sudo ip netns exec host4 sudo arp -s 192.168.222.3 32:fb:fa:c6:67:03
