package main

import (
	"loginsrv/handler"
	pb "loginsrv/proto"

	"github.com/go-micro/plugins/v4/logger/zap"
	"github.com/go-micro/plugins/v4/registry/kubernetes"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	uzap "go.uber.org/zap"
	uzapcore "go.uber.org/zap/zapcore"
)

var (
	service = "loginsrv"
	version = "latest"
)

func main() {
	// Create service
	srv := micro.NewService()

	//TODO: read level from env
	zapLogger, err := loggerWithLevel(logger.DebugLevel)
	if err != nil {
		logger.Errorf("could not create new logger: %s", err.Error())
	}
	logger.DefaultLogger = zapLogger

	srv.Init(
		micro.Name(service),
		micro.Version(version),
		micro.Logger(zapLogger),
		micro.Registry(kubernetes.NewRegistry()),
	)

	// Register handler
	if err := pb.RegisterLoginsrvHandler(srv.Server(), new(handler.Loginsrv)); err != nil {
		logger.Errorf("RegisterLoginsrvHandler: %s", err)
	}
	// Run service
	if err := srv.Run(); err != nil {
		logger.Errorf("could not run the service: %s", err)
	}
}

func loggerWithLevel(level logger.Level) (logger.Logger, error) {
	zapConfig := uzap.NewProductionConfig()
	zapConfig.EncoderConfig.EncodeTime = uzapcore.RFC3339NanoTimeEncoder
	zapConfig.Level.SetLevel(uzapcore.DebugLevel)

	loggerOpt := zap.WithConfig(zapConfig)
	logLevelOpt := logger.WithLevel(level)

	return zap.NewLogger(loggerOpt, logLevelOpt)
}
