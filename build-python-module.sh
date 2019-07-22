#!/usr/bin/env bash

python2.7 -m grpc_tools.protoc -Iprotobuf --python_out=python/push --grpc_python_out=python/push protobuf/push_service.proto
