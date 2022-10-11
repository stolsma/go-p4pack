#!/usr/bin/env bash
#
# Copyright 2022 - Sander Tolsma. All rights reserved
# SPDX-License-Identifier: Apache-2.0

# let alias work in non interactive bash shells!
shopt -s expand_aliases

# Builds are run as root in containers, no need for sudo
[ "$(id -u)" != '0' ] || alias sudo=

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
