package main

import (
	"context"
	"github.com/gorilla/handlers"
	runtime2 "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	gw "github.com/ueno-bst/grpc-gateway-skel/examples/grpc/skelton/v1"
	"github.com/ueno-bst/grpc-gateway-skel/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"net/http"
)

func main() {
	server, err := runtime.NewGateway(
		runtime.WithServer("0.0.0.0", 8081),
		runtime.WithBackend("0.0.0.0", 8080),
		runtime.WithEndpoint(
			gw.RegisterHelloWorldServiceHandlerFromEndpoint,
		),
		runtime.WithHealthCheckPathHandle("/ping/heartbeat"),
		runtime.WithStatusPathHandle("/ping/status"),
		runtime.WithCORS(
			handlers.AllowCredentials(),
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{http.MethodOptions, http.MethodGet}),
			handlers.AllowedHeaders([]string{"Authorization", "Content-Type", "Accept-Encoding", "Accept"}),
			handlers.MaxAge(300),
		),
		runtime.WithMetadata(
			runtime.PassThrowMeta([]string{"Cookie"}, runtime.RequestMeta),
			runtime.DeleteMeta([]string{"GRPC-Metadata-*"}, runtime.ResponseMeta),
		),
		runtime.WithAccessLogOutput("access.log"),
		runtime.WithErrorOutput("error.log"),
		runtime.WithHandler(
			runtime.CommonLogHandler,
			runtime.GzipCompressHandler,
			runtime.BrotliCompressHandler,
			//runtime.DeflateCompressHandler,
		),
		runtime.WithErrorHandler(
			runtime.ErrorHandle(codes.NotFound, func(ctx context.Context, mux *runtime2.ServeMux, w http.ResponseWriter, r *http.Request, s *status.Status) proto.Message {
				res := gw.ErrorResponse{
					Message: "messagesas",
					Random:  3232,
				}

				return &res
			})),
	)

	if err != nil {
		panic(err)
	}

	defer func() {
		server.Stop()
	}()

	if err := server.Start(); err != nil {
		panic(err)
	}
}
