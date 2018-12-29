package middleware

import (
	"context"

	"google.golang.org/grpc"
)

// UnaryMiddleWareHandler is an adapter to allow user to intercept the execution of a unary RPC on the server.
// parameters are defined as grpc.UnaryServerInterceptor.
type UnaryMiddleWareHandler func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo) (resp interface{}, err error)

// MiddleWare is an adapter to allow user to filter the execution of a unary RPC on the server
type MiddleWare func(handler UnaryMiddleWareHandler) UnaryMiddleWareHandler

// HandleWithGlobalMiddlewares provides a hook to intercept the execution of a unary RPC on the server
func HandleWithGlobalMiddlewares(interceptor grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// a transition from grpc.UnaryServerInterceptor to UnaryMiddleWareHandler
		next := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo) (resp interface{}, err error) {
			if interceptor != nil {
				return interceptor(ctx, req, info, handler)
			}

			return handler(ctx, req)
		}

		next = DebugRequestMiddleWare(next)
		return next(ctx, req, info)
	}
}
