#!/usr/bin/env bash

protoc -I protobuf/ \
--gogoslick_out=Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,\
plugins=grpc:server \
protobuf/push_service.proto

