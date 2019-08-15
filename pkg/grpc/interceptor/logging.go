package interceptor

import (
	"context"
	"encoding/json"
	"path"
	"time"

	"github.com/alibaba/pouch/pkg/log"
	"github.com/alibaba/pouch/pkg/randomid"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// ServerPayloadLoggingDecider is a user-provided function for deciding how to log the server-side
// request/response payloads
type ServerPayloadLoggingDecider func(ctx context.Context, fullMethodName string, servingObject interface{}) logrus.Level

// PayloadUnaryServerInterceptor returns a new unary server interceptors that logs the payloads of requests.
func PayloadUnaryServerInterceptor(decider ServerPayloadLoggingDecider) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// add request id for cri trace log
		ctx = log.NewContext(ctx, map[string]interface{}{"RequestID": randomid.Generate()[:10]})

		logLevel := decider(ctx, info.FullMethod, info.Server)

		service := path.Dir(info.FullMethod)[1:]
		method := path.Base(info.FullMethod)

		ctx = log.AddFields(ctx, map[string]interface{}{
			"grpc.service":    service,
			"grpc.method":     method,
			"grpc.start_time": time.Now().Format(time.RFC3339),
		})
		logProtoMessageAsJSON(ctx, req, "grpc.request.content", "grpc start", logLevel)
		resp, err := handler(ctx, req)
		if err == nil {
			logProtoMessageAsJSON(ctx, resp, "grpc.response.content", "grpc end", logLevel)
		} else {
			logProtoMessageAsJSON(ctx, resp, "grpc.response.content", "grpc failed "+err.Error(), logrus.ErrorLevel)
		}
		return resp, err
	}
}

func logProtoMessageAsJSON(ctx context.Context, pbMsg interface{}, key string, msg string, level logrus.Level) {
	b, _ := json.Marshal(pbMsg)
	entry := log.WithFields(ctx, map[string]interface{}{key: string(b)})

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
