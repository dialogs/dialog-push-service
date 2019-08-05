package service

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/dialogs/dialog-go-lib/service"
	httprouter "github.com/dialogs/dialog-go-lib/service/router"
	"github.com/dialogs/dialog-push-service/pkg/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Service struct {
	impl          *impl
	logger        *zap.Logger
	apiPort       string
	adminPort     string
	ctxDone       context.Context
	ctxDoneCancel func()
}

func New(cfg *viper.Viper, logger *zap.Logger) (*Service, error) {

	c, err := NewConfig(cfg)
	if err != nil {
		return nil, err
	}

	svcImpl, err := newImpl(c, logger)
	if err != nil {
		return nil, err
	}

	ctxDone, ctxDoneCancel := context.WithCancel(context.Background())

	return &Service{
		impl:          svcImpl,
		logger:        logger,
		apiPort:       c.ApiPort,
		adminPort:     c.AdminPort,
		ctxDone:       ctxDone,
		ctxDoneCancel: ctxDoneCancel,
	}, nil
}

func (s *Service) Close() error {
	s.ctxDoneCancel()
	return nil
}

func (s *Service) Run() error {

	var wgAdminSvc, wgApiSvc sync.WaitGroup

	adminRouter := httprouter.NewAdminRouter(Info())
	adminRouter.Handle("/metrics", promhttp.Handler())

	adminSvc := service.NewHTTP(adminRouter, time.Second)
	defer func() {
		if err := adminSvc.Close(); err != nil {
			s.logger.Error("close admin service", zap.Error(err))
		}
		wgAdminSvc.Wait()
	}()

	apiSvc := service.NewGRPC()
	defer func() {
		if err := apiSvc.Close(); err != nil {
			s.logger.Error("close API service", zap.Error(err))
		}
		wgApiSvc.Wait()
	}()

	retval := make(chan error, 3)

	wgAdminSvc.Add(1)
	go func() {
		defer wgAdminSvc.Done()
		address := net.JoinHostPort("0.0.0.0", s.adminPort)

		err := adminSvc.ListenAndServeAddr(address)
		if err != nil && err != http.ErrServerClosed {
			s.logger.Error("admin router closed", zap.Error(err))
		}
		retval <- err
	}()

	wgApiSvc.Add(1)
	go func() {
		defer wgApiSvc.Done()

		address := net.JoinHostPort("0.0.0.0", s.apiPort)

		apiSvc.RegisterService(func(grpcSvr *grpc.Server) {
			api.RegisterPushingServer(grpcSvr, s.impl)
		})
		err := apiSvc.ListenAndServeAddr(address)
		if err != nil && err != http.ErrServerClosed {
			s.logger.Error("admin router closed", zap.Error(err))
		}
		retval <- err
	}()

	go func() {
		<-s.ctxDone.Done()
		if err := apiSvc.Close(); err != nil {
			s.logger.Error("failed to close API service", zap.Error(err))
		}
	}()

	return <-retval
}
