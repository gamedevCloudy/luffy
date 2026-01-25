#!/usr/bin/env bash

test -d $PWD/builds || mkdir $PWD/builds

#platform=$(uname -s)
#arch=$(uname -m)
#
#if [[ "$platform" == "Linux" ]]; then
#  if [[ "$arch" == "aarch64" ]]; then
#    go build -ldflags="-s -w" -o builds/luffy-linux-aarch64
#    upx --best --lzma luffy.aarch64
#  elif [["$arch" == "riscv64" ]]; then
#    GOOS=linux GOARCH=riscv64 CGO_ENABLED=0 go build -ldflags="-s -w" -o builds/luffy-linux-rv64
#  else
#    go build -ldflags="-s -w" -o builds/luffy-linux-amd64
#    upx --best --lzma builds/luffy-linux-amd64
#  fi
#else
#  go build -o builds/luffy-macos-aarch64
#fi

build() {
  local os=$1
  local arch=$2
  GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -ldflags="-s -w" -o builds/$3
}

if [[ "$1" == "windows" ]]; then
  build $1 $2 luffy-$1-$2.exe
else
  build $1 $2 luffy-$1-$2
fi
