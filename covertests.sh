#!/usr/bin/env bash

echo "" > coverage.txt
code=0
for pkg in $(go list ./... | grep -v /vendor); do
    go test -race -coverprofile=profile.out -covermode=atomic $pkg || code=$?
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm -f profile.out
    fi
done
(exit $code)
