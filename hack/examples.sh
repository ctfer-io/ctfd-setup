#!/bin/bash

# Build binary
go build -cover -o ctfd-setup cmd/ctfd-setup/main.go
GOCOVERDIR=coverdir
mkdir "$GOCOVERDIR"

# Execute every examples
# WARNING: every '.admin' must be equal in order to reuse the CTFd instance
for dir in examples/*/; do
    if [[ -d "$dir" ]]; then
        (
            cd "$dir"
            GOCOVERDIR="../../$GOCOVERDIR" ../../ctfd-setup --url "$URL" --file .ctfd.yaml
        )
    fi
done

# Merge coverage data
go tool covdata textfmt "-i=$GOCOVERDIR" -o integration.out
sed -i '/^\//d' integration.out

# Remove traces
rm -rf "$GOCOVERDIR"
rm ctfd-setup
