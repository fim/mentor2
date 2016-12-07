#!/bin/bash
# build all arches
make GOOS=linux
make GOOS=windows
make GOOS=darwin
