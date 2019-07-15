package server

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"google.golang.org/grpc"
)

var opts struct {
	ConfigLocation   string `short:"c" long:"config" description:"Config file location" required:"true"`
	StupidUnusedArgs string `short:"g" long:"gelf" description:"Unusued"`
}

func (config *serverConfig) startGrpc() *grpc.Server {
	pushingServer := newPushingServer(config)
	logrusEntry := log.NewEntry(log.StandardLogger())
	grpcServer := grpc.NewServer(
		grpc_middleware.WithStreamServerChain(
			grpc_logrus.StreamServerInterceptor(logrusEntry),
			grpc_prometheus.StreamServerInterceptor,
		),
		//grpc.Creds(c)
	)
	grpc_prometheus.Register(grpcServer)
	RegisterPushingServer(grpcServer, pushingServer)
	return grpcServer
}

func StartServer() {
	var config *serverConfig
	var err error
	log.SetFormatter(&log.JSONFormatter{})
	if _, err = flags.ParseArgs(&opts, os.Args); err != nil {
		log.Fatalf("Failed to parse arguments: %s", err.Error())
	}
	if config, err = loadConfig(opts.ConfigLocation); err != nil {
		log.Fatalf("Failed to parse config: %s", err.Error())
	}
	grpc_logrus.ReplaceGrpcLogger(log.NewEntry(log.StandardLogger()))
	grpcServer := config.startGrpc()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GrpcPort))
	if err != nil {
		log.Fatalf("Failed to start gRPC server: %s", err.Error())
	}
	http.Handle("/metrics", prometheus.Handler())
	go func() {
		log.Infof("Started HTTP server at %d", config.HTTPPort)
		panic(http.ListenAndServe(fmt.Sprintf(":%d", config.HTTPPort), nil))
	}()
	log.Infof("Started gRPC server at %d", config.GrpcPort)
	panic(grpcServer.Serve(lis))
}
