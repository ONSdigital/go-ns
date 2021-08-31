#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/go-ns
  make lint
popd
