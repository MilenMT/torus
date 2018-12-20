#!/bin/sh

PATH=${PATH}:${GOBIN}
protoc --gogofaster_out=plugins=grpc:. -I. -I${GOPATH}/src -I${GOPATH}/src/github.com/gogo/protobuf/protobuf *.proto
