Dialog push service
===================

Dialog Push Service (DPS) is a tiny service for pushing remote notifications.

DPS consists of two sub-projects:

- [gRPC](http://www.grpc.io/) server that does actual delivery
- Scala module for [Dialog Server](https://dlg.im) which provides Scala bindings to gRPC server

Getting Started
---------------

You need to install go, protobuf and gogo-protobuf extensions:
```
brew install go
brew install protobuf
go get github.com/gogo/protobuf/proto
go get github.com/gogo/protobuf/protoc-gen-gogoslick
go get github.com/gogo/protobuf/gogoproto
```
You also need to make sure the `protoc-gen-gogoslick` is in the `$PATH`.

License
-------

This software is available under the [Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0.html)
