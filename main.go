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

	//TODO:fix log level
	// devConfig.Level.SetLevel(uzapcore.DebugLevel)

	loggerOpt := zap.WithConfig(devConfig)
	zapLogger, err := zap.NewLogger(loggerOpt)
	if err != nil {
		fmt.Println(err)
	}
	logger.DefaultLogger = zapLogger

	srv.Init(
		micro.Name(service),
		micro.Version(version),
		micro.Logger(zapLogger),
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
