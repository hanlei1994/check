#!/usr/bin/env bash

set -e
# build bin file
export GO111MODULE=on
export GOPROXY=https://goproxy.io
go build -v
# make my dir
mkdir -p  $BUILD_ROOT
mv check  $BUILD_ROOT
mv conf.toml  $BUILD_ROOT
