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

readonly DEFAULT_UBUNTU_VERSION=20.04
readonly DEFAULT_IMAGE_NAME=ghcr.io/stolsma/p4c

export UBUNTU_VERSION=${UBUNTU_VERSION:-$DEFAULT_UBUNTU_VERSION}
export IMAGE_NAME=${IMAGE_NAME:-$DEFAULT_IMAGE_NAME}

readonly DPDK_TAG=$IMAGE_NAME:latest

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

pushd "$THIS_DIR" > /dev/null
  docker pull "ubuntu:$UBUNTU_VERSION"
  docker build \
    -t "$DPDK_TAG" \
    --build-arg UBUNTU_VERSION="$UBUNTU_VERSION" \
    .
    
  if $PUSH; then
    docker push "$DPDK_TAG"
  fi
popd > /dev/null