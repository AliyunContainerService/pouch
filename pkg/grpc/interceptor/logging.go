package interceptor

import (
	"context"
	"path"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// ServerPayloadLoggingDecider is a user-provided function for deciding how to log the server-side
// request/response payloads
type ServerPayloadLoggingDecider func(ctx context.Context, fullMethodName string, servingObject interface{}) logrus.Level

// PayloadUnaryServerInterceptor returns a new unary server interceptors that logs the payloads of requests.
func PayloadUnaryServerInterceptor(entry *logrus.Entry, decider ServerPayloadLoggingDecider) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logLevel := decider(ctx, info.FullMethod, info.Server)

		service := path.Dir(info.FullMethod)[1:]
		method := path.Base(info.FullMethod)

		logEntry := entry.WithFields(logrus.Fields{
			"grpc.service":    service,
			"grpc.method":     method,
			"grpc.start_time": time.Now().Format(time.RFC3339),
		})
		logProtoMessageAsJSON(logEntry, req, "grpc.request.content", "grpc start", logLevel)
		resp, err := handler(ctx, req)
		if err == nil {
			logProtoMessageAsJSON(logEntry, resp, "grpc.response.content", "grpc end", logLevel)
		} else {
			logProtoMessageAsJSON(logEntry, resp, "grpc.response.content", "grpc failed "+err.Error(), logrus.ErrorLevel)
		}
		return resp, err
	}
}

func logProtoMessageAsJSON(entry *logrus.Entry, pbMsg interface{}, key string, msg string, level logrus.Level) {
	p, ok := pbMsg.(proto.Message)
	if !ok {
		return
	}
	entry = entry.WithField(key, &jsonpbMarshalleble{p})

	switch level {
	case logrus.DebugLevel:
		entry.Debug(msg)
	case logrus.InfoLevel:
		entry.Info(msg)
	case logrus.WarnLevel:
		entry.Warn(msg)
	case logrus.ErrorLevel:
		entry.Error(msg)
	case logrus.FatalLevel:
		entry.Fatal(msg)
	case logrus.PanicLevel:
		entry.Panic(msg)
	default:
		entry.Debug(msg)
	}
}

type jsonpbMarshalleble struct {
	proto.Message
}
