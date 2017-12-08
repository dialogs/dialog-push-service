FROM golang:latest

RUN curl https://glide.sh/get | sh

ADD src/server /go/src/server
WORKDIR /go/src/server

RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN go build
RUN go install
RUN ls -la /go/bin
RUN ls -la /go/src/server

ENTRYPOINT ["/go/bin/server"]
