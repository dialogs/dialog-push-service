#!/usr/bin/env bash

python2.7 -m grpc_tools.protoc -Isrc/protobuf --python_out=src/python/push --grpc_python_out=src/python/push src/protobuf/push_service.proto
