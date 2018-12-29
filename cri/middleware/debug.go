package middleware

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	// defaultRecordTimeOut defines a certain period of time, if a cri request spent
	// more than it, a log will be recorded to reminder it. the unit of measurement is millisecond.
	defaultRecordTimeOut = 500
)

// DebugRequestMiddleWare add log for cri request
func DebugRequestMiddleWare(handler UnaryMiddleWareHandler) UnaryMiddleWareHandler {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo) (resp interface{}, err error) {
		logrus.WithField("method", info.FullMethod).Debugf("Cri request: %v", req)

		start := time.Now()
		defer func() {
			d := time.Since(start) / (time.Millisecond)
			if d > defaultRecordTimeOut {
				logrus.WithField("method", info.FullMethod).Infof("End of Calling Cri request: %v, costs %d ms", req, d)
			}
		}()

		return handler(ctx, req, info)
	}
}
