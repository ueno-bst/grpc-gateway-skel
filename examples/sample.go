package main

import (
	"github.com/gorilla/handlers"
	"github.com/ueno-bst/grpc-gateway-skel/runtime"
	"net/http"
)

func main() {
	server := runtime.NewGateway(
		runtime.WithHealthCheckPathHandle("/ping/heartbeat"),
		runtime.WithStatusPathHandle("/ping/status"),
		runtime.WithCORS(
			handlers.AllowCredentials(),
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{http.MethodOptions, http.MethodGet}),
			handlers.AllowedHeaders([]string{"Authorization", "Content-Type", "Accept-Encoding", "Accept"}),
			handlers.MaxAge(300),
		),
		runtime.WithHandler(
			runtime.GzipCompressHandler,
			runtime.BrotliCompressHandler,
		),
	)

	defer func() {
		server.Stop()
	}()

	if err := server.Start(); err != nil {
		panic(err)
	}
}
