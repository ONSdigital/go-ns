#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/go-ns
  make audit
popd