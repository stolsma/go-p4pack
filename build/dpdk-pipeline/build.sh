#!/bin/bash
#Copyright (C) 2022 Sander Tolsma
#SPDX-License-Identifier: Apache-2.0

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

if [ -z "$UBUNTU_VERSION" ]; then
    UBUNTU_VERSION=20.04
fi

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

pushd $THIS_DIR > /dev/null
  docker pull "ubuntu:$UBUNTU_VERSION"
  docker build -t stolsma/dpdk-infra --build-arg BASE_IMG=ubuntu:"$UBUNTU_VERSION" .
  if $PUSH; then
    docker push stolsma/dpdk-infra
  fi
popd > /dev/null