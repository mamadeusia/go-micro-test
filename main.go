package main

import (
	"loginsrv/handler"
	pb "loginsrv/proto"

	"github.com/go-micro/plugins/v4/logger/zap"
	"go-micro.dev/v4"
	uzap "go.uber.org/zap"
	uzapcore "go.uber.org/zap/zapcore"
)

var (
	service = "loginsrv"
	version = "latest"
)

func main() {
	// Create service

	// logger.Info("sdfsdfs")
	srv := micro.NewService()
	// logger, err := zap.NewLogger()
	// if err != nil {
	// 	logger.Logf(2, "sdfsdf")

	// }
	// logger.Logf(2, "sdfsdf")
	prodEC := uzap.NewProductionEncoderConfig()
	prodEC.EncodeTime = uzapcore.RFC3339TimeEncoder

	zapConfig := uzap.Config{
		Level:             2,
		Development:       true,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          &uzap.SamplingConfig{},
		Encoding:          "",
		EncoderConfig:     prodEC,
		OutputPaths:       []string{},
		ErrorOutputPaths:  []string{},
		InitialFields:     map[string]interface{}{},
	}

	loggerOpt := zap.WithConfig(zapConfig)

	logger, _ := zap.NewLogger(loggerOpt)

	srv.Init(
		micro.Name(service),
		micro.Version(version),
		micro.Logger(logger),
	)

	// Register handler
	if err := pb.RegisterLoginsrvHandler(srv.Server(), new(handler.Loginsrv)); err != nil {
		logger.Logf(2, "Err", err)
	}
	logger.Log(logger.DebugLevel, "sdfsdfsd")
	// Run service
	if err := srv.Run(); err != nil {
		logger.Logf(2, "Err", err)
	}
}
