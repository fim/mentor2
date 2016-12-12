#!/bin/bash
# build all arches
make GOOS=linux
make GOOS=windows
make GOOS=darwin
make GOOS=freebsd
find bin -type f -executable -exec sha256sum {} \; > SHA256
