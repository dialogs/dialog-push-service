FROM golang:1.12 as builder

WORKDIR $GOPATH/src/github.com/dialogs/dialog-push-service

ADD server server
ADD main.go main.go
ADD go.mod go.mod
ADD go.sum go.sum

RUN GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -ldflags="-w -s" -o /dialog-push-service main.go

FROM debian:stretch-slim

WORKDIR /

COPY --from=builder /dialog-push-service /dialog-push-service

CMD ["/dialog-push-service"]
