Dialog push service
===================

Dialog Push Service (DPS) is a tiny service for pushing remote notifications.

DPS consists of two sub-projects:

- [gRPC](http://www.grpc.io/) server that does actual delivery
- Scala module for [Dialog Server](https://dlg.im) which provides Scala bindings to gRPC server

Getting Started
---------------

You need to install go, protobuf and gogo-protobuf extensions:
```bash
brew install go
brew install protobuf
go get github.com/gogo/protobuf/proto
go get github.com/gogo/protobuf/protoc-gen-gogoslick
go get github.com/gogo/protobuf/gogoproto
```
You also need to make sure the `protoc-gen-gogoslick` is in the `$PATH`.

## Test environment:

1. download iOS certificate in *PEM* format
2. create environment variable __APPLE_PUSH_CERTIFICATE__ with path to *PEM*
3. download [service-account.json](https://console.firebase.google.com/project/_/settings/serviceaccounts/adminsdk)
4. create environment variable __GOOGLE_APPLICATION_CREDENTIALS__ with path to *service-account.json*
5. create devices tokens file. File format:
```json
{
  "android": "<token>",
  "ios": "<token>"
}
```
1. create environment variable __PUSH_DEVICES__ with path to *devices tokens file*


## Client application

### Android

Build and run:
- copy [google-services.json](https://console.firebase.google.com/project/_/settings/general/android:com.example.push) to /pkg/test/app/android/app
- build application
- run android emulator

License
-------

This software is available under the [Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0.html)