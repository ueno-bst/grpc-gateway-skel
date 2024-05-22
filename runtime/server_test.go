package runtime

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestWithServer(t *testing.T) {
	tests := []struct {
		name string
		host string
		port uint
		want GatewayOptionFunc
	}{
		{
			"Empty host & zero port",
			"",
			0,
			func(opt *GatewayOption) { opt.server = ServerInfo{"", 0} },
		},
		{
			"Non-empty host & zero port",
			"example.com",
			0,
			func(opt *GatewayOption) { opt.server = ServerInfo{"example.com", 0} },
		},
		{
			"Empty host & non-zero port",
			"",
			8080,
			func(opt *GatewayOption) { opt.server = ServerInfo{"", 8080} },
		},
		{
			"Non-empty host & non-zero port",
			"example.com",
			8080,
			func(opt *GatewayOption) { opt.server = ServerInfo{"example.com", 8080} },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithServer(tt.host, tt.port); !reflect.DeepEqual(reflect.ValueOf(got).Pointer(), reflect.ValueOf(tt.want).Pointer()) {
				t.Errorf("WithServer() = %v, want %v", reflect.ValueOf(got).Pointer(), reflect.ValueOf(tt.want).Pointer())
			}
		})
	}
}

func TestServerInfo_Valid(t *testing.T) {
	values := []ServerInfo{
		{"256.220.30.1", 80},
		{"a.b.c.d", 80},
		{"0::xdsa", 80},
		{"0::0", 65536},
	}

	for _, value := range values {
		err := value.Valid()

		assert.Errorf(t, err, "Allowing unacceptable IPs %+v", value)
	}
}

func TestWithHandler(t *testing.T) {
	server := NewGateway(
		WithServer("0.0.0.0", 8080),
		WithHandler(GzipCompressHandler),
		WithHandler(BrotliCompressHandler),
	)

	assert.Contains(t, server.handlers, GzipCompressHandler)
	assert.Contains(t, server.handlers, BrotliCompressHandler)
}
