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
ARG DPDKINFRA_HOME=/dpdkinfra

ENV GOPATH /go
ENV CGO_CFLAGS_ALLOW .*
ENV CGO_LDFLAGS_ALLOW .*
ENV PKG_CONFIG_PATH /usr/local/share/pkgconfig:$PKG_CONFIG_PATH

# hadolint ignore=DL3003
RUN mkdir --parents $DPDKINFRA_HOME \
  && git clone https://github.com/stolsma/go-p4pack.git \
  && cd go-p4pack/ \
  && go mod tidy \
  && go build -v ./cmd/dpdkinfra/... \
  && mv ./dpdkinfra $DPDKINFRA_HOME \
  && mv ./examples $DPDKINFRA_HOME \
  && mv ./README.md $DPDKINFRA_HOME \
  && cd .. \
  && rm -Rf go-p4pack

WORKDIR ${DPDKINFRA_HOME}
