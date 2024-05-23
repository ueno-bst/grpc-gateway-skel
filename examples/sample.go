package main

import (
	"github.com/gorilla/handlers"
	"github.com/ueno-bst/grpc-gateway-skel/runtime"
	"net/http"
)

func main() {
	server, err := runtime.NewGateway(
		runtime.WithHealthCheckPathHandle("/ping/heartbeat"),
		runtime.WithStatusPathHandle("/ping/status"),
		runtime.WithCORS(
			handlers.AllowCredentials(),
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{http.MethodOptions, http.MethodGet}),
			handlers.AllowedHeaders([]string{"Authorization", "Content-Type", "Accept-Encoding", "Accept"}),
			handlers.MaxAge(300),
		),
		//runtime.WithAccessLogOutput("access.log"),
		//runtime.WithErrorOutput("error.log"),
		runtime.WithHandler(
			runtime.CommonLogHandler,
			runtime.DeflateCompressHandler,
			runtime.GzipCompressHandler,
			runtime.BrotliCompressHandler,
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
