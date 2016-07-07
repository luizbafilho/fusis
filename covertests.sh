#!/usr/bin/env bash

echo "" > coverage.txt
code=0
for pkg in $(go list ./... | grep -v /vendor); do
    sudo -E go test -race -coverprofile=profile.out -covermode=atomic $pkg || code=$?
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done
(exit $code)
