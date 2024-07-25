package runtime

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
)

type ErrorHandleCallback = func(ctx context.Context, mux *runtime.ServeMux, w http.ResponseWriter, r *http.Request, s *status.Status) proto.Message

func WithErrorHandler(handles ...ErrorHandleReturn) GatewayOptionFunc {
	return func(opt *GatewayOption) {
		for _, handle := range handles {
			handle(opt)
		}

		opt.muxOpts = append(opt.muxOpts, errorCapture(*opt))
	}
}

type ErrorHandleReturn = func(opt *GatewayOption)

func ErrorHandle(code codes.Code, callback ErrorHandleCallback) ErrorHandleReturn {
	return func(opt *GatewayOption) {
		opt.errors[code] = callback
	}
}

func errorCapture(opt GatewayOption) runtime.ServeMuxOption {
	const fallback = `{"code": 13, "message": "failed to marshal error message"}`

	return runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshal runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
		s, ok := status.FromError(err)

		if !ok {
			s = status.New(codes.Unknown, err.Error())
		}

		if callback := opt.errors[s.Code()]; callback != nil {
			if r := callback(ctx, mux, w, r, s); r != nil {
				contentType := marshal.ContentType(r)
				w.Header().Set("Content-Type", contentType)

				buf, err := marshal.Marshal(r)
				if err != nil {
					grpclog.Errorf("Failed to marshal error message %q: %v", s, err)
					w.WriteHeader(http.StatusInternalServerError)
					if _, err := io.WriteString(w, fallback); err != nil {
						grpclog.Errorf("Failed to write response: %v", err)
					}
					return
				}

				w.WriteHeader(runtime.HTTPStatusFromCode(s.Code()))

				if _, err := w.Write(buf); err != nil {
					grpclog.Errorf("Failed to write response: %v", err)
				}

				return
			}
		}

		runtime.DefaultHTTPErrorHandler(ctx, mux, marshal, w, r, err)
	})
}
