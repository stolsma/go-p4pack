ARG UBUNTU_VERSION

FROM ubuntu:$UBUNTU_VERSION

ARG DPDK_VERSION
ARG DPDK_STUFF=dpdk-$DPDK_VERSION.tar.xz
ARG DPDK_HOME=/dpdk

WORKDIR /
COPY install.sh /scripts/
COPY apply-patches.sh /scripts/
COPY patches /scripts/patches/

WORKDIR /scripts
RUN ./install.sh && rm -Rf /root/.ccache
