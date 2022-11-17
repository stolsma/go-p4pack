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

# Create host0
sudo ip netns add host1
sudo ip link set sw1 netns host1
sudo ip netns exec host1 sudo ip link set lo up
sudo ip netns exec host1 sudo ip link set sw1 address 32:fb:fa:c6:67:01
sudo ip netns exec host1 sudo ip -4 addr add 192.168.222.1/24 dev sw1
sudo ip netns exec host1 sudo ip link set sw1 up
sudo ip netns exec host1 sudo arp -s 192.168.222.2 32:fb:fa:c6:67:02

# Create host1
sudo ip netns add host2
sudo ip link set sw2 netns host2
sudo ip netns exec host2 sudo ip link set lo up
sudo ip netns exec host2 sudo ip link set sw2 address 32:fb:fa:c6:67:02
sudo ip netns exec host2 sudo ip -4 addr add 192.168.222.2/24 dev sw2
sudo ip netns exec host2 sudo ip link set sw2 up
sudo ip netns exec host2 sudo arp -s 192.168.222.1 32:fb:fa:c6:67:01
