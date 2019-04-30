FROM golang:latest

ADD src/server /go/src/server
WORKDIR /go/src/server

RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure -v -vendor-only
RUN go build
RUN go install
RUN ls -la /go/bin
RUN ls -la /go/src/server

ENTRYPOINT ["/go/bin/server"]
