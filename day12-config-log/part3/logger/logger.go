package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

var Logger *zap.Logger

// InitLogger 初始化日志
func InitLogger(level, filePath string, maxSize, maxBackups, maxAge int) error {
	// 日志级别
	var zapLevel zapcore.Level
	switch level {
	case "debug":
	zapLevel = zapcore.DebugLevel
	case "info":
	zapLevel = zapcore.InfoLevel
	case "warn":
	zapLevel = zapcore.WarnLevel
	case "error":
	zapLevel = zapcore.ErrorLevel
	default:
	zapLevel = zapcore.InfoLevel
	}

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,  // 彩色输出
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 文件输出
	fileWriter := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    maxSize,    // MB
		MaxBackups: maxBackups,
		MaxAge:     maxAge,     // days
		Compress:   true,
	}

	// 控制台输出
	consoleWriter := zapcore.AddSync(os.Stdout)

	// 创建 Core
	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(fileWriter),
			zapLevel,
		),
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			consoleWriter,
			zapLevel,
		),
	)

	// 创建 Logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

func Sync(){
	if Logger != nil {
		Logger.Sync()
	}
}