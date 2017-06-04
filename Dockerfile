FROM golang:latest

RUN curl https://glide.sh/get | sh

ADD src/server /go/src/server
WORKDIR /go/src/server

RUN ls -la
RUN pwd
RUN glide install
RUN go build
RUN go install
RUN ls -la /go/bin
RUN ls -la /go/src/server

ENTRYPOINT ["/go/bin/server"]
