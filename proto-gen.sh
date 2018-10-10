#!/usr/bin/env bash

#protoc -I src/protobuf/ src/protobuf/push_service.proto --go_out=plugins=grpc:src/server
protoc -I src/protobuf/ \
--gogoslick_out=Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,\
plugins=grpc:src/server \
src/protobuf/push_service.proto

