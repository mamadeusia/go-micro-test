package main

import (
	"loginsrv/handler"
	pb "loginsrv/proto"

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

	logger.Info("sdfsdfs")
	srv := micro.NewService()
	// logger, err := zap.NewLogger()
	// if err != nil {
	// 	logger.Logf(2, "sdfsdf")

	// }
	// logger.Logf(2, "sdfsdf")
	// var opt zap.Options
	prodEC := uzap.NewProductionEncoderConfig()
	prodEC.EncodeTime = uzapcore.RFC3339TimeEncoder
	// opt = append(opt, zap.WithEncoderConfig(prodEC), zap.WithNamespace("test"))
	// 	// zap.WithConfig(uzap.NewDevelopmentConfig())) // zap.WithConfig()

	// opts := zap.Options{
	// 	zap.WithEncoderConfig(prodEC),

	// }

	srv.Init(
		micro.Name(service),
		micro.Version(version),
		// micro.Logger(logger),
		// micro.
	)

	// Register handler
	if err := pb.RegisterLoginsrvHandler(srv.Server(), new(handler.Loginsrv)); err != nil {
		logger.Fatal(err)
	}
	logger.Log(logger.DebugLevel, "sdfsdfsd")
	// Run service
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
