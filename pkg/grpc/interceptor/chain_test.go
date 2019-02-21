package interceptor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestChainUnaryServer(t *testing.T) {
	type key int

	var (
		serviceName  = "Foo.UnaryMethod"
		theUnaryInfo = &grpc.UnaryServerInfo{FullMethod: serviceName}

		keyFirst    key = 1
		keySecond   key = 2
		valueFirst      = "I'm first"
		valueSecond     = "I'm second"

		input  = "input"
		result = "result"
	)

	first := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		assert.Equal(t, ctx.Value(keyFirst), nil, "there should be no first value")
		assert.Equal(t, info, theUnaryInfo)
		return handler(context.WithValue(ctx, keyFirst, valueFirst), req)
	}

	second := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		assert.Equal(t, ctx.Value(keyFirst), valueFirst, "there should be first value")
		assert.Equal(t, ctx.Value(keySecond), nil, "there should be no second value")
		assert.Equal(t, info, theUnaryInfo)
		return handler(context.WithValue(ctx, keySecond, valueSecond), req)
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		assert.Equal(t, req, input, "check the input")
		assert.Equal(t, ctx.Value(keyFirst), valueFirst, "there should be first value")
		assert.Equal(t, ctx.Value(keySecond), valueSecond, "there should be second value")
		return result, nil
	}

	output, err := chainUnaryServer(first, second)(context.TODO(), input, theUnaryInfo, handler)
	assert.Equal(t, err, nil)
	assert.Equal(t, output, result)
}
