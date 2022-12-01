#!/usr/bin/env bash
#
# Copyright 2022 - Sander Tolsma. All rights reserved
# SPDX-License-Identifier: Apache-2.0

sudo mkdir /mnt/huge
sudo mount -t hugetlbfs nodev /mnt/huge
sudo sysctl -w vm.nr_hugepages=256

docker run \
		--name go-p4pack \
		--rm \
		--cap-add ALL \
		--privileged \
		-v "${PWD}":/dummy \
		-v /mnt/huge:/mnt/huge \
		-v /sys/bus/pci/devices:/sys/bus/pci/devices \
		-v /sys/devices/system/node:/sys/devices/system/node \
		-v /dev:/dev \
		-p 9339:9339 \
		-p 9559:9559 \
		-p 2222:2222 \
		--entrypoint /bin/bash \
		-it ghcr.io/stolsma/go-p4pack:latest 
