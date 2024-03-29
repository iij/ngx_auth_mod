#!/bin/bash
set -eu
export GO111MODULE=on
OUT_DIR=$(pwd)/bin

echo "output: ${OUT_DIR}"
mkdir -p "${OUT_DIR}"
ls -1 src/ngx_auth/exec | while read d ; do
  echo -n "build ${d}: "
  (
    cd "src/ngx_auth/exec/${d}" || exit 1
    go build -o "${OUT_DIR}" -ldflags '-s -w' || exit
  ) || continue
  echo done
done
