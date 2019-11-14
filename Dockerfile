FROM golang:1.13-stretch as builder

ARG COMMIT=
ARG RELEASE=
ARG PROJECT=github.com/dialogs/dialog-push-service

WORKDIR $GOPATH/src/$PROJECT

ADD pkg pkg
ADD service service
ADD main.go main.go
ADD go.mod go.mod
ADD go.sum go.sum

RUN CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -ldflags "-s -w \
    -X ${PROJECT}/pkg/info.Commit=${COMMIT} \
    -X ${PROJECT}/pkg/info.Version=${RELEASE} \
    -X ${PROJECT}/pkg/info.GoVersion=$(go version| sed -e 's/ /_/g') \
    -X ${PROJECT}/pkg/info.BuildDate=$(date -u '+%Y-%m-%d_%H:%M:%S')" \
    -race -v \
    -o /push-server main.go

FROM debian:stretch-slim

COPY --from=builder /push-server /usr/local/bin/push-server

RUN apt update -y && \
    apt install -y ca-certificates && \
    update-ca-certificates

USER 1000

CMD ["push-server"]
