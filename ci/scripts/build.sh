#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/go-ns
  make build
popd