package main

import (
	"fmt"
	"loginsrv/handler"
	pb "loginsrv/proto"

	"github.com/go-micro/plugins/v4/logger/zap"
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

	devConfig := uzap.NewProductionConfig()
	devConfig.EncoderConfig.EncodeTime = uzapcore.RFC3339NanoTimeEncoder

	loggerOpt := zap.WithConfig(devConfig)

	zapLogger, err := zap.NewLogger(loggerOpt)
	if err != nil {
		fmt.Println("======", err)
	}

	srv.Init(
		micro.Name(service),
		micro.Version(version),
		micro.Logger(zapLogger),
	)

	zapLogger.Logf(logger.ErrorLevel, "Test log", "test 2")

	// Register handler
	if err := pb.RegisterLoginsrvHandler(srv.Server(), new(handler.Loginsrv)); err != nil {
		zapLogger.Logf(logger.ErrorLevel, "Err", err)
	}
	// Run service
	if err := srv.Run(); err != nil {
		zapLogger.Logf(logger.ErrorLevel, "Err", err)
	}
}
