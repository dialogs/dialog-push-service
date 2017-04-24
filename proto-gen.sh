#!/usr/bin/env bash

#protoc -I src/protobuf/ src/protobuf/push_service.proto --go_out=plugins=grpc:src/server
protoc -I src/protobuf/ src/protobuf/push_service.proto --gogoslick_out=plugins=grpc:src/server
