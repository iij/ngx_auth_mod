#!/bin/bash
set -eu
export GO111MODULE=on
export GOBIN=$(pwd)/bin

echo "GOBIN: ${GOBIN}"
ls -1 src/ngx_auth/exec | while read d ; do
  echo -n "install ${d}: "
  (
    cd "src/ngx_auth/exec/${d}" || exit 1
    go install -ldflags '-s -w' || exit
  ) || continue
  echo done
done
