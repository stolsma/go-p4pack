#!/usr/bin/env bash

set -eo pipefail

echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections

apt-get -qq update
apt-get -qq install --no-install-recommends \
  build-essential \
  libtool \
  clang \
  gcc \
  g++ \
  ccache \
  pkg-config \
  autoconf \
  autoconf-archive \
  automake \
  libtool \
  libxml2-dev \
  libdw-dev \
  libbsd-dev \
  libpcap-dev \
  libibverbs-dev \
  libcrypto++-dev \
  libfdt-dev \
  libjansson-dev \
  libarchive-dev \
  zlib1g-dev \
  iproute2 \
  ca-certificates \
  libnuma-dev \
  python3-pip \
  python3-pyelftools \
  python3-setuptools \
  python3-wheel\
  wget \
  xz-utils \
  git
pip3 install \
  meson \
  ninja
apt-get -qq clean
rm --recursive --force /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Set all arguments from default if not given
readonly DEFAULT_DPDK_VERSION=22.07
DPDK_VERSION=${DPDK_VERSION:-$DEFAULT_DPDK_VERSION}
readonly DEFAULT_DPDK_STUFF=dpdk-$DPDK_VERSION.tar.xz
DPDK_STUFF=${DPDK_STUFF:-$DEFAULT_DPDK_STUFF}
readonly DEFAULT_DPDK_HOME=/dpdk
DPDK_HOME=${DPDK_HOME:-$DEFAULT_DPDK_HOME}

echo *******************************************************************************
echo Getting DPDK source files!
echo "DPDK version: $DPDK_VERSION"
echo "DPDK Tar    : $DPDK_STUFF"
echo "DPDK Home   : $DPDK_HOME"
echo

# Download and extract the requested DPDK source version
wget --quiet http://fast.dpdk.org/rel/"$DPDK_STUFF"
mkdir --parents "$DPDK_HOME"
tar --extract --file="$DPDK_STUFF" --directory="$DPDK_HOME" --strip-components 1
rm --force "$DPDK_STUFF"

echo
echo *******************************************************************************

# Apply the needed patches on the DPDK source files
./apply-patches.sh

# Build DPDK
pushd "$DPDK_HOME" > /dev/null
  meson build
  cd build
  ninja
  ninja install
  ldconfig
popd > /dev/null

# And remove all DPDK source and build files
rm --recursive --force "$DPDK_HOME"

echo *******************************************************************************
echo DPDK Build and Install Ready!
echo *******************************************************************************
