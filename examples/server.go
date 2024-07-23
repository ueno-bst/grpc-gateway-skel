package main

import (
	"github.com/gorilla/handlers"
	gw "github.com/ueno-bst/grpc-gateway-skel/examples/grpc/skelton/v1"
	"github.com/ueno-bst/grpc-gateway-skel/runtime"
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
			runtime.BrotliCompressHandler,
			//runtime.GzipCompressHandler,
			//runtime.DeflateCompressHandler,
		),
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
