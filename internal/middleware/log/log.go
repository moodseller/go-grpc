package log

import (
	grpclogrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"time"
)

func UnaryServerInterceptor(logger *logrus.Logger, logSkipper func(method string) bool) grpc.UnaryServerInterceptor {
	logrusEntry := logrus.NewEntry(logger)
	opts := []grpclogrus.Option{
		grpclogrus.WithDurationField(func(duration time.Duration) (key string, value interface{}) {
			return "grpc.time_ns", duration.Nanoseconds()
		}),
		grpclogrus.WithDecider(func(fullMethodName string, err error) bool {
			return !logSkipper(fullMethodName)
		}),
	}
	return grpclogrus.UnaryServerInterceptor(logrusEntry, opts...)
}

func StreamServerInterceptor(logger *logrus.Logger) grpc.StreamServerInterceptor {
	logrusEntry := logrus.NewEntry(logger)
	opts := []grpclogrus.Option{
		grpclogrus.WithDurationField(func(duration time.Duration) (key string, value interface{}) {
			return "grpc.time_ns", duration.Nanoseconds()
		}),
	}
	return grpclogrus.StreamServerInterceptor(logrusEntry, opts...)
}

// ReplaceGrpcLogger replaces grpc system logger with custom
// logrus logger and sets log level to error only.
func ReplaceGrpcLogger(logger *logrus.Logger) {
	logger.SetLevel(logrus.ErrorLevel)
	systemLogrusEntry := logrus.NewEntry(logger).WithField("system", "grpc")
	grpclog.SetLoggerV2(&grpcSystemLogger{systemLogrusEntry})
}

type grpcSystemLogger struct {
	*logrus.Entry
}

func (l *grpcSystemLogger) V(v int) bool {
	return v <= int(l.Level)
}
