#!/usr/bin/env bash

HERE=$(dirname $(readlink "$0"))
go run "$HERE"/go/ai/jenai.go "$@"
