#!/bin/bash
#Copyright (C) 2022 Sander Tolsma
#SPDX-License-Identifier: Apache-2.0

set -e

# shellcheck source=build/dpdk-pipeline/scripts/os_ver_details.sh
. scripts/os_ver_details.sh
get_os_ver_details

if [ -z "$1" ]
then
   echo "-Missing mandatory arguments:"
   echo " - Usage: ./build.sh <WORKDIR> "
   return 1
fi

WORKDIR=$1
cd "${WORKDIR}" || exit

echo "Removing p4pipeline directory if it already exists"
if [ -d "p4pipeline" ]; then rm -Rf p4pipeline; fi
mkdir "$1/p4pipeline" && cd "$1/p4pipeline" || exit
#..Setting Environment Variables..#
echo "Exporting Environment Variables....."
export SDE="${PWD}"
export SDE_INSTALL="$SDE/install"
mkdir "$SDE_INSTALL" || exit

#...Package Config Path...#
if [ "${OS}" = "Ubuntu" ]  || [ "${VER}" = "20.04" ] ; then
    export PKG_CONFIG_PATH=${SDE_INSTALL}/lib/x86_64-linux-gnu/pkgconfig
else
    export PKG_CONFIG_PATH=${SDE_INSTALL}/lib64/pkgconfig
fi

#..Runtime Path...#
export LD_LIBRARY_PATH=$SDE_INSTALL/lib
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$SDE_INSTALL/lib64
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/lib64
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/lib

echo "SDE environment variable"
echo "$SDE"
echo "$SDE_INSTALL"
echo "$PKG_CONFIG_PATH"

#Read the number of CPUs in a system and derive the NUM threads
get_num_cores
echo "Number of Parallel threads used: $NUM_THREADS ..."
echo ""

cd "$SDE" || exit
echo "Removing p4vswitch repository if it already exists"
if [ -d "p4vswitch" ]; then rm -Rf p4vswitch; fi
echo "Compiling dpdk-pipeline"
git clone https://github.com/stolsma/p4vswitch.git p4vswitch
pushd "$SDE/p4vswitch"
git checkout main
git submodule update --init --recursive

# compile dpdk-pipeline
cd src
./apply_patch.sh > /dev/null 2>&1
cd dpdk_src
meson -Dexamples=pipeline build
cd build
ninja
ninja install
cd ../..

# create special dpdk-pipeline version
echo "Compiling dpdk_infra"
make -C infra install_dir=/root

set +e