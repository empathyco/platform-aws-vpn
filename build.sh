#!/bin/bash

set -eu

BUILDDIR=$(mktemp -d)
trap 'rm -r ${BUILDDIR}' EXIT

ARTIFACTSDIR=./artifacts
mkdir -p "${ARTIFACTSDIR}"

build_lambdas() {
  for CMD in ./cmd/*; do
    [ -e "${CMD}/main.go" ] || continue
    CMD_BASE=$(basename "${CMD}")
    BINARY_NAME=${CMD_BASE/#lambda-/vpn-}

    go build -o "${BUILDDIR}/${BINARY_NAME}" "${CMD}"
    zip -j "${ARTIFACTSDIR}/${BINARY_NAME}.zip" "${BUILDDIR}/${BINARY_NAME}"
  done
}

build_frontend() {
  npm --prefix ./frontend install
  npm --prefix ./frontend run build
  tar -czf ./artifacts/frontend.tar.gz -C ./frontend/dist .
}

build_all() {
  build_lambdas
  build_frontend
}

case $1 in
lambdas)
  build_lambdas
  ;;
frontend)
  build_frontend
  ;;
all)
  build_all
  ;;
*)
  printf "Unknown command %s\n" "$1"
  ;;
esac
