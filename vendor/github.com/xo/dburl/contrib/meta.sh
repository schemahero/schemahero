#!/bin/bash

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../)

pushd $SRC &> /dev/null

gometalinter \
  --disable=aligncheck \
  --disable=goconst \
  --enable=misspell \
  --enable=gofmt \
  --deadline=100s \
  --cyclo-over=15 \
  --sort=path \
  --exclude='^url\.go.*function Parse\(\).*\(gocyclo\)$' \
  ./...

popd &> /dev/null
