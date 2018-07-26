package stream

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func TestWriteError(t *testing.T) {
	for _, tt := range []struct {
		err  error
		code int
	}{
		{
			fmt.Errorf("error that not from grpc package"),
			http.StatusInternalServerError,
		},
		{
			ErrorStreamingDisabled("methodNotExist"),
			http.StatusNotFound,
		},
		{
			ErrorTooManyInFlight(),
			http.StatusTooManyRequests,
		},
	} {
		res := httptest.NewRecorder()
		WriteError(tt.err, res)
		if res.Code != tt.code {
			t.Fatalf("unexpected code in http response")
		}
		if !reflect.DeepEqual([]byte(tt.err.Error()), res.Body.Bytes()) {
			t.Fatalf("unexpected content in the body of http response")
		}
		if grpc.Code(tt.err) == codes.ResourceExhausted {
			if res.Header().Get("Retry-After") != strconv.Itoa(int(CacheTTL.Seconds())) {
				t.Fatalf("the Retry-After field of http header is not expected when the error is resource exhausted")
			}
		}
	}
}
