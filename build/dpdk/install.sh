#!/usr/bin/env bash

# let alias work in non interactive bash shells!
shopt -s expand_aliases

# Builds are run as root in containers, no need for sudo
[ "$(id -u)" != '0' ] || alias sudo=

sudo apt-get -y update
sudo DEBIAN_FRONTEND=noninteractive apt-get -y install --no-install-recommends \
  build-essential \
  libtool \
  ccache \
  pkg-config \
  autoconf \
  automake \
  clang \
  gcc \
  g++ \
  autoconf-archive \
  libxml2-dev \
  libdw-dev \
  libbsd-dev \
  libpcap-dev \
  libibverbs-dev \
  libcrypto++-dev \
  libfdt-dev \
  libjansson-dev \
  libarchive-dev \
  libnuma-dev \
  python3-pip \
  python3-pyelftools \
  python3-setuptools \
  python3-wheel\
  ca-certificates \
  iproute2 \
  wget \
  xz-utils \
  git \
  net-tools \
  iputils-ping \
  sudo
sudo pip3 install \
  meson \
  ninja
sudo apt-get -qq clean
sudo rm --recursive --force /var/lib/apt/lists/* /tmp/* /var/tmp/*

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
./apply-patches.sh "$DPDK_HOME"

# Build DPDK
pushd "$DPDK_HOME" > /dev/null || exit
  meson -Dcpu_instruction_set=generic build
  cd build || exit
  sed -e 's/RTE_CPUFLAG_AVX512BW\,//' -i rte_build_config.h
	sed -e 's/RTE_CPUFLAG_AVX512CD\,//' -i rte_build_config.h
	sed -e 's/RTE_CPUFLAG_AVX512DQ\,//' -i rte_build_config.h
	sed -e 's/RTE_CPUFLAG_AVX512F\,//' -i rte_build_config.h
	sed -e 's/RTE_CPUFLAG_AVX512VL\,//' -i rte_build_config.h
  ninja
  sudo ninja install
  sudo ldconfig
popd > /dev/null || exit

# And remove all DPDK source and build files
rm --recursive --force "$DPDK_HOME"

echo *******************************************************************************
echo DPDK Build and Install Ready!
echo *******************************************************************************
