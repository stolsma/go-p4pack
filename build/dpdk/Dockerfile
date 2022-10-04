ARG UBUNTU_VERSION

FROM ubuntu:$UBUNTU_VERSION

ARG DPDK_VERSION
ARG DPDK_STUFF=dpdk-$DPDK_VERSION.tar.xz
ARG DPDK_HOME=/dpdk

WORKDIR /
COPY install.sh /
COPY apply-patches.sh /
COPY patches /patches/

RUN /install.sh

#RUN set -o pipefail && echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections
#
#RUN apt-get -qq update \
#  && apt-get -qq install --no-install-recommends \
#       build-essential \
#       libtool \
#       clang \
#       gcc \
#       g++ \
#       ccache \
#       pkg-config \
#       autoconf \
#       autoconf-archive \
#       automake \
#       libtool \
#       libxml2-dev \
#       libdw-dev \
#       libbsd-dev \
#       libpcap-dev \
#       libibverbs-dev \
#       libcrypto++-dev \
#       libfdt-dev \
#       libjansson-dev \
#       libarchive-dev \
#       zlib1g-dev \
#       iproute2 \
#       ca-certificates \
#       libnuma-dev \
#       python3-pip \
#      python3-pyelftools \
#       python3-setuptools \
#       python3-wheel\
#       wget \
#       xz-utils \
#       git \
#  && pip3 install \
#       meson \
#       ninja \
#  && apt-get -qq clean \
#  && rm --recursive --force /var/lib/apt/lists/* /tmp/* /var/tmp/* \
#  && wget --quiet http://fast.dpdk.org/rel/$DPDK_STUFF \
#  && mkdir --parents $DPDK_HOME \
#  && tar --extract --file=$DPDK_STUFF --directory=$DPDK_HOME --strip-components 1 \
#  && rm --force $DPDK_STUFF \
#  && /apply-patches.sh \
#  && cd $DPDK_HOME \
#  && meson build \
#  && cd build \
#  && ninja \
#  && ninja install \
#  && ldconfig \
#  && cd ~ \
#  && rm --recursive --force $DPDK_HOME
