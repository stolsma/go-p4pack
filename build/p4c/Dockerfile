ARG UBUNTU_VERSION=20.04
FROM ubuntu:${UBUNTU_VERSION} as p4c-build

ENV P4C_BUILD_DEPS bison \
                   build-essential \
                   libboost-dev \
                   libboost-iostreams-dev \
                   ca-certificates \
                   cmake \
                   libfl-dev \
                   libgc-dev \
                   git \
                   libgmp-dev \
                   protobuf-compiler \
                   libprotobuf-dev \
                   python3

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y install --no-install-recommends $P4C_BUILD_DEPS && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

ARG P4C_URL=https://github.com/p4lang/p4c.git
ARG P4C_REF=main

WORKDIR /

RUN git clone $P4C_URL

WORKDIR /p4c/

RUN git checkout $P4C_REF
RUN git submodule update --init --recursive

#COPY apply-patches.sh /
#COPY patches /patches/

#RUN /apply-patches.sh

WORKDIR /p4c/build

RUN cmake -DCMAKE_INSTALL_PREFIX=/usr/local -DCMAKE_CXX_FLAGS:STRING=-O3 -DENABLE_GTESTS=OFF -DENABLE_P4C_GRAPHS=OFF .. && \
    make && \
    make DESTDIR=/output install

FROM ubuntu:${UBUNTU_VERSION}
LABEL maintainer="Sander Tolsma <github@tolsma.net>"
LABEL description="A image for the p4lang/p4c compiler based on Ubuntu"

# runtime dependencies
# python is required by the p4c driver
# python3-minimal is missing shlex
ENV P4C_RUNTIME_DEPS libboost-iostreams1.71.0 \
                     libgc1c2 \
                     cpp \
                     libgmp10 \
                     python3

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y install --no-install-recommends $P4C_RUNTIME_DEPS && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY --from=p4c-build /output/usr/local /usr/local/