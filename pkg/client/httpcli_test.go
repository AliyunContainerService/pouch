package client

import (
	"crypto/tls"
	"crypto/x509"
	"reflect"
	"testing"

	"github.com/alibaba/pouch/pkg/serializer"
	"github.com/go-resty/resty"
)

func Test_isPtr(t *testing.T) {
	type args struct {
		i interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPtr(tt.args.i); got != tt.want {
				t.Errorf("isPtr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseTLSConfig(t *testing.T) {
	type args struct {
		ca   []byte
		cert []byte
		key  []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *tls.Config
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTLSConfig(tt.args.ca, tt.args.cert, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTLSConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTLSConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_certpool(t *testing.T) {
	type args struct {
		pem []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *x509.CertPool
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := certpool(tt.args.pem)
			if (err != nil) != tt.wantErr {
				t.Errorf("certpool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("certpool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClientNew(t *testing.T) {
	tests := []struct {
		name string
		want *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HTTPClientNew(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPClientNew() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClient_verb(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	type args struct {
		method string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			if got := c.verb(tt.args.method); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPClient.verb() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClient_TLSConfig(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	type args struct {
		tlsc *tls.Config
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			if got := c.TLSConfig(tt.args.tlsc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPClient.TLSConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClient_TLS(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	type args struct {
		ca   []byte
		cert []byte
		key  []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *HTTPClient
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			got, err := c.TLS(tt.args.ca, tt.args.cert, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPClient.TLS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPClient.TLS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPUT(t *testing.T) {
	tests := []struct {
		name string
		want *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PUT(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PUT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGET(t *testing.T) {
	tests := []struct {
		name string
		want *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GET(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GET() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPOST(t *testing.T) {
	tests := []struct {
		name string
		want *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := POST(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("POST() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDELETE(t *testing.T) {
	tests := []struct {
		name string
		want *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DELETE(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DELETE() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClient_PUT(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	tests := []struct {
		name   string
		fields fields
		want   *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			if got := c.PUT(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPClient.PUT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClient_GET(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	tests := []struct {
		name   string
		fields fields
		want   *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			if got := c.GET(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPClient.GET() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClient_POST(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	tests := []struct {
		name   string
		fields fields
		want   *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			if got := c.POST(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPClient.POST() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClient_DELETE(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	tests := []struct {
		name   string
		fields fields
		want   *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			if got := c.DELETE(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPClient.DELETE() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClient_Err(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			if err := c.Err(); (err != nil) != tt.wantErr {
				t.Errorf("HTTPClient.Err() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHTTPClient_Method(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	type args struct {
		method string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			if got := c.Method(tt.args.method); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPClient.Method() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClient_URL(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	type args struct {
		rawurl string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			if got := c.URL(tt.args.rawurl); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPClient.URL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClient_SetHeader(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			if got := c.SetHeader(tt.args.key, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPClient.SetHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClient_Body(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	type args struct {
		obj serializer.Object
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			if got := c.Body(tt.args.obj); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPClient.Body() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClient_JSONBody(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	type args struct {
		obj serializer.Object
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			if got := c.JSONBody(tt.args.obj); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPClient.JSONBody() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClient_Query(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	type args struct {
		obj serializer.Object
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			if got := c.Query(tt.args.obj); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPClient.Query() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClient_Do(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	tests := []struct {
		name   string
		fields fields
		want   *HTTPClient
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			if got := c.Do(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPClient.Do() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClient_RespCodeEqual(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	type args struct {
		code int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			if got := c.RespCodeEqual(tt.args.code); got != tt.want {
				t.Errorf("HTTPClient.RespCodeEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClient_StatusCode(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			if got := c.StatusCode(); got != tt.want {
				t.Errorf("HTTPClient.StatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClient_Into(t *testing.T) {
	type fields struct {
		cli    *resty.Client
		req    *resty.Request
		err    error
		resp   *resty.Response
		method string
		urls   string
		code   int
	}
	type args struct {
		obj serializer.Object
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HTTPClient{
				cli:    tt.fields.cli,
				req:    tt.fields.req,
				err:    tt.fields.err,
				resp:   tt.fields.resp,
				method: tt.fields.method,
				urls:   tt.fields.urls,
				code:   tt.fields.code,
			}
			if err := c.Into(tt.args.obj); (err != nil) != tt.wantErr {
				t.Errorf("HTTPClient.Into() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
