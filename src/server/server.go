package server

import (
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/jessevdk/go-flags"
	"github.com/mwitkow/go-grpc-middleware"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
)

var opts struct {
	ConfigLocation string `short:"c" long:"config" description:"Config file location" required:"true"`
}

func (config *serverConfig) startGrpc() *grpc.Server {
	pushingServer := newPushingServer(config)
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(grpc_prometheus.StreamServerInterceptor)),
	)
	grpc_prometheus.Register(grpcServer)
	RegisterPushingServer(grpcServer, pushingServer)
	return grpcServer
}

func StartServer() {
	var config *serverConfig
	var err error
	if _, err = flags.ParseArgs(&opts, os.Args); err != nil {
		grpclog.Fatalf("Failed to parse arguments: %s", err.Error())
	}
	if config, err = loadConfig(opts.ConfigLocation); err != nil {
		grpclog.Fatalf("Failed to parse config: %s", err.Error())
	}
	grpcServer := config.startGrpc()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GrpcPort))
	if err != nil {
		grpclog.Fatalf("Failed to start gRPC server: %s", err.Error())
	}
	prometheus.MustRegister(fcmIOHistogram, apnsIOHistogram)
	http.Handle("/metrics", prometheus.Handler())
	go func() {
		grpclog.Printf("Started HTTP server at port %d", config.HTTPPort)
		panic(http.ListenAndServe(fmt.Sprintf(":%d", config.HTTPPort), nil))
	}()
	grpclog.Printf("Started gRPC server at port %d", config.GrpcPort)
	panic(grpcServer.Serve(lis))
}
