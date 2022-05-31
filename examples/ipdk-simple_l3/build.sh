#!/bin/bash
#
# Copyright 2022 - Sander Tolsma. All rights reserved
# SPDX-License-Identifier: Apache-2.0

docker run --rm -u 1000:1000 -v "${PWD}":/p4ccode -w /p4ccode stolsma/p4c-all:latest /bin/bash -c "p4c-dpdk -o simple_l3.spec --p4runtime-files p4info.proto.txt simple_l3.p4"