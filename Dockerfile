FROM golang:1.11

ADD . /go/src/dialog-push-service
WORKDIR /go/src/dialog-push-service

ENV GO111MODULE=on
RUN go install
RUN ls -la /go/bin
RUN ls -la /go/src/dialog-push-service

ENTRYPOINT ["/go/bin/dialog-push-service"]
