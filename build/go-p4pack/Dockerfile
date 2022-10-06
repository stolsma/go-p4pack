ARG DPDK_TAG

# create the go-build environment
# hadolint ignore=DL3006
FROM $DPDK_TAG as go-build
ARG GOLANG_VERSION

# Install Go
RUN wget --progress=dot:giga https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz \
  && tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz \
  && rm -f go${GOLANG_VERSION}.linux-amd64.tar.gz

ENV PATH "$PATH:/usr/local/go/bin"

# Build the go app #as dpdkinfra-build
FROM go-build 
ARG GOP4PACK_HOME=/go-p4pack

ENV GOPATH /go
ENV CGO_CFLAGS_ALLOW .*
ENV CGO_LDFLAGS_ALLOW .*
ENV PKG_CONFIG_PATH /usr/local/share/pkgconfig:$PKG_CONFIG_PATH

# hadolint ignore=DL3003
RUN mkdir --parents $GOP4PACK_HOME/build \
  && cd "$GOP4PACK_HOME/build" \
  && git clone https://github.com/stolsma/go-p4pack.git \
  && cd go-p4pack/ \
  && go mod tidy \
  && go build -v ./cmd/dpdkinfra/... \
  && mv ./dpdkinfra $GOP4PACK_HOME \
  && mv ./examples $GOP4PACK_HOME \
  && mv ./README.md $GOP4PACK_HOME \
  && cd $GOP4PACK_HOME \
  && rm -Rf "$GOP4PACK_HOME/build"

WORKDIR ${GOP4PACK_HOME}
