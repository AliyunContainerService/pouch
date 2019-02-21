package interceptor

import (
	"context"

	"google.golang.org/grpc"
)

// WithUnaryServerChain creates a single interceptor out of a chain of many interceptors.
func WithUnaryServerChain(interceptors ...grpc.UnaryServerInterceptor) grpc.ServerOption {
	return grpc.UnaryInterceptor(chainUnaryServer(interceptors...))
}

// chainUnaryServer creates a single interceptor out of a chain of many interceptors.
//
// Execution is done in left-to-right order.
func chainUnaryServer(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		buildChain := func(current grpc.UnaryServerInterceptor, next grpc.UnaryHandler) grpc.UnaryHandler {
			return func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return current(currentCtx, currentReq, info, next)
			}
		}

		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = buildChain(interceptors[i], chain)
		}
		return chain(ctx, req)
	}
}
