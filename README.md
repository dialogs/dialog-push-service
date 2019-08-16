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

## Config

### [Legacy FCM HTTP](https://firebase.google.com/docs/cloud-messaging/http-server-ref)

```yaml
google:
  - project-id: <string>
    key: <string>
    retries: <number>
    timeout: <string>
    nop-mode: <boolean>
    workers: <number>
    allow-alerts: <boolean>
    sandbox: <boolean>
```
properties:
- project-id - identificator of the provider
- [key](https://firebase.google.com/docs/cloud-messaging/auth-server#authorize_legacy_protocol_send_requests)
- retries - count retries by server error
- timeout - time duration. Example: 1s, 2m
- nop-mode - if the option is set to true, the message will not be sent
- workers - count workers for sending. By default the value is equal count of processors.
- allow-alerts - enabled alerting messages for converter protobuf push message to a notification message
- sandbox - if the option is set to true, the message will not be actually sent. Instead FCM performs all the necessary validations, and emulates the send operation

### [FCM HTTP v1](https://firebase.google.com/docs/cloud-messaging/concept-options)

```yaml
fcm:
  - project-id: <string>
    service-account: <string>
    retries: <number>
    timeout: <string>
    nop-mode: <boolean>
    workers: <number>
    allow-alerts: <boolean>
    sandbox: <boolean>
```
properties:
- project-id - identificator of the provider
- [service-account](https://console.firebase.google.com/project/_/settings/serviceaccounts/adminsdk)
- retries - count retries by server error
- timeout - time duration. Example: 1s, 2m
- nop-mode - if the option is set to true, the message will not be sent
- workers - count workers for sending. By default the value is equal count of processors.
- allow-alerts - enabled alerting messages for converter protobuf push message to a notification message
- sandbox - if the option is set to true, the message will not be actually sent. Instead FCM performs all the necessary validations, and emulates the send operation

### [APNS](https://developer.apple.com/library/archive/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/APNSOverview.html#//apple_ref/doc/uid/TP40008194-CH8-SW1)

```yaml
apple:
  - project-id: <string>
    pem: <string>
    retries: <number>
    timeout: <string>
    nop-mode: <boolean>
    workers: <number>
    allow-alerts: <boolean>
    sandbox: <boolean>
```
properties:
- project-id - identifier of the provider
- pem - path to tls certificate in pem format
- retries - count retries by server error
- timeout - time duration. Example: 1s, 2m
- nop-mode - if the option is set to true, the message will not be sent
- workers - count workers for sending. By default the value is equal count of processors.
- allow-alerts - enabled alerting messages for converter protobuf push message to a notification message
- topic - the [topic](https://developer.apple.com/library/archive/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/CommunicatingwithAPNs.html#//apple_ref/doc/uid/TP40008194-CH11-SW1) of the remote notification, which is typically the bundle ID for your ap
- sound - sound of the alerting message


## Test environment

1. download iOS certificate in *PEM* format
2. create environment variable __APPLE_PUSH_CERTIFICATE__ with path to *PEM*
3. download [service-account.json](https://console.firebase.google.com/project/_/settings/serviceaccounts/adminsdk)
4. create environment variable __GOOGLE_APPLICATION_CREDENTIALS__ with path to *service-account.json*
5. copy [server key](https://console.firebase.google.com/project/_/settings/cloudmessaging/android:com.example.push)
6. save *server key* in a file. File format:
```json
{
  "key":"<server key>"
}
```
7. create environment variable __GOOGLE_LEGACY_APPLICATION_CREDENTIALS__ with path to server key file
8. create devices tokens file. File format:
```json
{
  "android": "<token>",
  "ios": "<token>"
}
```
9. create environment variable __PUSH_DEVICES__ with path to *devices tokens file*


## Client application

### Android

Build and run:
- copy [google-services.json](https://console.firebase.google.com/project/_/settings/general/android:com.example.push) to /pkg/test/app/android/app
- build application
- run android emulator

License
-------

This software is available under the [Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0.html)