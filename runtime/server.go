package runtime

import (
	"context"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"net/http"
)

type GatewayHandler func(mux http.Handler) http.Handler

type GatewayEndpoint func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error

type PathHandler struct {
	method  string
	path    string
	handler runtime.HandlerFunc
}

type ServerInfo struct {
	host string
	port uint
}

type ServerTLS struct {
	cert string
	key  string
}

func (s ServerInfo) Valid() error {
	if s.port < 0 || s.port > 65535 {
		return fmt.Errorf("invalid port number: %d", s.port)
	}

	ip := net.ParseIP(s.host)

	if ip == nil {
		return fmt.Errorf("invalid ip address: %s", s.host)
	}

	return nil
}

func (s ServerInfo) ToString() (string, error) {
	if err := s.Valid(); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", s.host, s.port), nil
}

type GatewayOption struct {
	server    ServerInfo
	tls       *ServerTLS
	backend   ServerInfo
	endpoints []GatewayEndpoint
	handlers  []GatewayHandler
	paths     []PathHandler
	silent    bool
	ctx       *context.Context
}

type GatewayOptionFunc func(*GatewayOption)

type GatewayServer interface {
	Start() error
	Stop() GatewayServer
	Restart() error
}

func NewGateway(opts ...GatewayOptionFunc) GatewayOption {
	o := GatewayOption{
		server:    ServerInfo{"0.0.0.0", 8081},
		backend:   ServerInfo{"127.0.0.1", 8080},
		endpoints: []GatewayEndpoint{},
		handlers:  []GatewayHandler{},
		paths:     []PathHandler{},
		silent:    false,
	}

	for _, opt := range opts {
		opt(&o)
	}

	return o
}

// WithServer is a GatewayOptionFunc that sets the host and port for the server.
// It takes the host and port as parameters and sets them in the ServerInfo struct.
// When used as an argument for GatewayServer.WithOptions or GatewayServer.WithDefaultOptions,
// it sets the host and port for the server to listen on.
// The ServerInfo struct represents the host and port information of the server.
//
// Example usage:
//
//	server := NewGateway(
//	    WithServer("0.0.0.0", 8080),
//	)
func WithServer(host string, port uint) GatewayOptionFunc {
	return func(opt *GatewayOption) {
		opt.server = ServerInfo{host, port}
	}
}

// WithTLS is a GatewayOptionFunc that sets the TLS certificate and key for the server.
// It takes the paths of the certificate and key files as parameters.
// When used as an argument for GatewayServer.WithOptions or GatewayServer.WithDefaultOptions,
// it sets the TLS configuration for the server to enable secure communication.
func WithTLS(cert, key string) GatewayOptionFunc {
	return func(opt *GatewayOption) {
		opt.tls = &ServerTLS{cert, key}
	}
}

func WithSilent(silent bool) GatewayOptionFunc {
	return func(opt *GatewayOption) {
		opt.silent = silent
	}
}

// WithBackend is a GatewayOptionFunc that sets the backend host and port.
// It takes the host and port as parameters and sets them in the GatewayOption struct.
// When used as an argument for GatewayServer.WithOptions or GatewayServer.WithDefaultOptions,
// it sets the backend configuration to connect to the specified host and port.
func WithBackend(host string, port uint) GatewayOptionFunc {
	return func(opt *GatewayOption) {
		opt.backend = ServerInfo{host, port}
	}
}

// WithEndpoint is a GatewayOptionFunc that appends the provided GatewayEndpoints to the
// endpoints field in the GatewayOption struct. It takes GatewayEndpoints as parameters
// and appends them to the existing endpoints in the GatewayOption struct. When used as
// an argument for GatewayServer.WithOptions or GatewayServer.WithDefaultOptions, it adds
// the provided endpoints to the server configuration.
func WithEndpoint(endpoints ...GatewayEndpoint) GatewayOptionFunc {
	return func(opt *GatewayOption) {
		opt.endpoints = append(opt.endpoints, endpoints...)
	}
}

func WithPathHandle(method string, path string, handler runtime.HandlerFunc) GatewayOptionFunc {
	return func(opt *GatewayOption) {
		opt.paths = append(opt.paths, PathHandler{method, path, handler})
	}
}

// WithHandler is a GatewayOptionFunc that adds one or more GatewayHandler functions to the GatewayOption struct.
// It takes a variadic parameter of type GatewayHandler and appends them to the handlers slice in the GatewayOption struct.
// When used as an argument for GatewayServer.WithOptions or GatewayServer.WithDefaultOptions,
// it adds the specified handlers to be applied to incoming requests.
// The GatewayHandler type is a function that takes an http.Handler and returns an http.Handler,
// allowing for custom middleware to be applied to the request chain.
//
// Example usage:
//
//	server := NewGateway(
//	    WithHandler(GzipCompressHandler),
//	    WithHandler(BrotliCompressHandler),
//	)
func WithHandler(handlers ...GatewayHandler) GatewayOptionFunc {
	return func(opt *GatewayOption) {
		opt.handlers = append(opt.handlers, handlers...)
	}
}

// WithCORS is a GatewayOptionFunc that adds the CORS handler to the GatewayOption struct.
// It takes a variadic parameter of type handlers.CORSOption and appends the CORS handler to the handlers slice in the GatewayOption struct.
// When used as an argument for GatewayServer.WithOptions or GatewayServer.WithDefaultOptions,
// it adds the CORS handler to be applied to incoming requests.
//
// Example usage:
//
//	server := NewGateway(
//	    WithCORS(
//			handlers.AllowCredentials(),
//			handlers.AllowedOrigins([]string{"*"}),
//			handlers.AllowedMethods([]string{http.MethodOptions, http.MethodGet}),
//			handlers.AllowedHeaders([]string{"Authorization", "Content-Type", "Accept-Encoding", "Accept"}),
//			handlers.MaxAge(300),
//		),
//	)
func WithCORS(option ...handlers.CORSOption) GatewayOptionFunc {
	return func(opt *GatewayOption) {
		opt.handlers = append(opt.handlers, handlers.CORS(option...))
	}
}

func (o *GatewayOption) attachHandler(mux http.Handler) http.Handler {
	if o.handlers != nil {
		for _, handler := range o.handlers {
			mux = handler(mux)
		}
	}

	return mux
}

func (o *GatewayOption) attachEndpoint(ctx context.Context, mux *runtime.ServeMux, opts []grpc.DialOption) error {
	host, err := o.backend.ToString()

	if err != nil {
		return err
	}

	if o.endpoints != nil {

		for _, endpoint := range o.endpoints {
			if err := endpoint(ctx, mux, host, opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func (o *GatewayOption) attachPathHandle(mux *runtime.ServeMux) error {
	if o.paths != nil {
		for _, path := range o.paths {
			if err := mux.HandlePath(path.method, path.path, path.handler); err != nil {
				return err
			}
		}
	}

	return nil
}

// Start starts the server by creating a new context with cancel function and setting it to o.ctx.
// It registers the gRPC server backend, attaches the endpoints, and starts the server using http.ListenAndServe or http.ListenAndServeTLS.
// It returns an error if any operation fails.
func (o *GatewayOption) Start() error {
	tls := o.tls

	host, err := o.server.ToString()

	if err != nil {
		return err
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer func() {
		fmt.Printf("Gateway server stopping on %s\n", host)
		cancel()
	}()

	o.ctx = &ctx

	// Register gRPC server backend
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()

	if err := o.attachPathHandle(mux); err != nil {
		return err
	}

	handler := o.attachHandler(mux)

	var credential credentials.TransportCredentials

	if tls != nil {
		if c, err := credentials.NewClientTLSFromFile(tls.cert, ""); err != nil {
			return err
		} else {
			credential = c
		}
	} else {
		credential = insecure.NewCredentials()
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credential),
	}

	if err := o.attachEndpoint(ctx, mux, opts); err != nil {
		return err
	}

	fmt.Printf("Gateway server starting on %s\n", host)
	if tls != nil {
		return http.ListenAndServeTLS(host, tls.cert, tls.key, handler)
	}

	return http.ListenAndServe(host, handler)
}

// Stop stops the server by canceling the context and setting it to nil.
// It returns the GatewayServer instance for method chaining.
func (o *GatewayOption) Stop() GatewayServer {
	if o.ctx != nil {
		_, cancel := context.WithCancel(*o.ctx)

		cancel()

		o.ctx = nil
	}

	return o
}

// Restart restarts the server by first stopping it using the Stop method and then starting it using the Start method.
func (o *GatewayOption) Restart() error {
	return o.Stop().Start()
}
