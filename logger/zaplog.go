package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func CreateLogger() *zap.Logger {
    stdout := zapcore.AddSync(os.Stdout)

    file := zapcore.AddSync(&lumberjack.Logger{
        Filename:   "logs/mini-docker.log",
        MaxSize:    10, // megabytes
        MaxBackups: 3,
        MaxAge:     7, // days
    })

    level := zap.NewAtomicLevelAt(zap.InfoLevel)

    productionCfg := zap.NewProductionEncoderConfig()
    productionCfg.TimeKey = "timestamp"
    productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

    developmentCfg := zap.NewDevelopmentEncoderConfig()
    developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

    consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
    fileEncoder := zapcore.NewJSONEncoder(productionCfg)

    core := zapcore.NewTee(
        zapcore.NewCore(consoleEncoder, stdout, level),
        zapcore.NewCore(fileEncoder, file, level),
    )

    return zap.New(core, zap.AddCaller())
}
