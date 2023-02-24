#!/usr/bin/env bash

set -eo pipefail

function echoerr {
    >&2 echo "$@"
}

function print_usage {
    echoerr "Usage: $0 [--push]"
}

PUSH=false
while [[ $# -gt 0 ]]
do
  key="$1"
  case $key in
    --push)
    PUSH=true
    shift
    ;;
    -h|--help)
    print_usage
    exit 0
    ;;
    *)    # unknown option
    echoerr "Unknown option $1"
    exit 1
    ;;
  esac
done

readonly DEFAULT_DPDK_TAG=ghcr.io/stolsma/dpdk-base:dpdk-22.07-ubuntu-20.04
readonly DEFAULT_IMAGE_NAME=ghcr.io/stolsma/go-p4pack
readonly DEFAULT_GOP4PACK_VERSION=1.0rc2
readonly DEFAULT_GOLANG_VERSION=1.20

export DPDK_TAG=${DPDK_TAG:-$DEFAULT_DPDK_TAG}
export IMAGE_NAME=${IMAGE_NAME:-$DEFAULT_IMAGE_NAME}
export GOP4PACK_VERSION=${GOP4PACK_VERSION:-$DEFAULT_GOP4PACK_VERSION}
export GOLANG_VERSION=${GOLANG_VERSION:-$DEFAULT_GOLANG_VERSION}

readonly INFRA_TAG=$IMAGE_NAME:$GOP4PACK_VERSION
readonly INFRA_TAG_LATEST=$IMAGE_NAME:latest

echo Golang version: "$GOLANG_VERSION"
echo DPDK source Tag: "$DPDK_TAG"
echo GOP4PACK version: "$GOP4PACK_VERSION"
echo GOP4PACK Tag: "$INFRA_TAG"

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

pushd "$THIS_DIR" > /dev/null
  docker build \
    --no-cache \
    --build-arg DPDK_TAG="$DPDK_TAG" \
    --build-arg GOLANG_VERSION="$GOLANG_VERSION" \
    --tag "$INFRA_TAG" \
    --tag "$INFRA_TAG_LATEST" \
    .
  if $PUSH; then
    docker push -a "$IMAGE_NAME"
  fi
popd > /dev/null
