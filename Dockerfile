FROM golang:latest

RUN go get github.com/spf13/viper\
 && go get github.com/jessevdk/go-flags\
 && go get github.com/edganiukov/fcm\
 && go get github.com/golang/protobuf/proto\
 && go get golang.org/x/net/http2\
 && go get golang.org/x/crypto/pkcs12\
 && go get github.com/sideshow/apns2\
 && go get google.golang.org/grpc\
 && go get github.com/mwitkow/go-grpc-middleware\
 && go get github.com/grpc-ecosystem/go-grpc-prometheus\
 && go get github.com/prometheus/client_golang/prometheus\
 && go get github.com/gogo/protobuf/proto\
 && go get -u go.uber.org/zap\
 && go get github.com/gogo/protobuf/sortkeys

COPY src /go/src

RUN go install server\
 && go build app

ENTRYPOINT ["/go/app"]
