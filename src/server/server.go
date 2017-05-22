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
	"go.uber.org/zap"
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
	logger, err := zap.NewProduction()
	defer logger.Sync()
	if err != nil {
		grpclog.Fatalf("Error initializing ZAP logger", err)
	}
	if _, err = flags.ParseArgs(&opts, os.Args); err != nil {
		logger.Fatal("Failed to parse arguments.", zap.Error(err))
	}
	if config, err = loadConfig(opts.ConfigLocation, logger); err != nil {
		logger.Fatal("Failed to parse config.", zap.Error(err))
	}
	grpcServer := config.startGrpc()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GrpcPort))
	if err != nil {
		logger.Fatal("Failed to start gRPC server.", zap.Error(err))
	}
	prometheus.MustRegister(fcmIOHistogram, apnsIOHistogram)
	http.Handle("/metrics", prometheus.Handler())
	go func() {
		logger.Info("Started HTTP server.", zap.Uint16("port", config.HTTPPort))
		panic(http.ListenAndServe(fmt.Sprintf(":%d", config.HTTPPort), nil))
	}()
	logger.Info("Started gRPC server.", zap.Uint16("port", config.GrpcPort))
	panic(grpcServer.Serve(lis))
}
