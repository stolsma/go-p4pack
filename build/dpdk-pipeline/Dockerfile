# Copyright 2022 - Sander Tolsma. All rights reserved
# SPDX-License-Identifier: BSD-3-Clause

ARG BASE_IMG=${BASE_IMG}

FROM ${BASE_IMG} AS base
ARG BASE_IMG

SHELL ["/bin/bash", "-c"]

RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections

# Installing dependent packages required for DPDK Build
RUN apt-get update

RUN apt-get install -y apt-utils \
    git \
    libtool \
    clang \
    gcc \
    g++ \
    autoconf \
    automake \
    autoconf-archive \
    libconfig++-dev \
    libgc-dev \
    unifdef \
    libffi-dev \
    libboost-iostreams-dev \
    libboost-graph-dev \
    llvm \
    pkg-config \
    flex libfl-dev \
    zlib1g-dev \
    python \
    python3-setuptools \
    python3-pip \
    python3-wheel \
    python3-cffi \
    sudo && \
    if [ "$BASE_IMG" = "ubuntu:18.04" ] ; then \
       apt-get -y install python-pip; \
    else \
       apt-get -y install pip; \
    fi && \
    apt-get -y clean all

# Installing all PYTHON packages
RUN python3 -m pip install --no-cache-dir --upgrade pip && \
    python3 -m pip install --no-cache-dir cmake>=3.15.0 \
    pyelftools \
    meson==0.59.4 \
    ninja>=1.8.2

# Build dpdk-infra driver
FROM base AS build-dpdk-infra
WORKDIR /root
COPY ./scripts scripts
RUN  /root/scripts/build.sh /root

# Create lean dpdk-infra image based on Ubuntu base image
FROM ${BASE_IMG} AS dpdk-infra
LABEL maintainer="github@tolsma.net"
SHELL ["/bin/bash", "-c"]
RUN apt-get update && apt-get install -y libatomic1 libbsd0 && apt-get -y clean all
COPY --from=build-dpdk-infra /root/dpdk_infra ./
COPY ./scripts scripts
