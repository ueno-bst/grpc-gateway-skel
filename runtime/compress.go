package runtime

import (
	"compress/flate"
	"compress/gzip"
	"github.com/andybalholm/brotli"
	"net/http"
	"strings"
)

// gzipResponseWriter is a type that wraps an http.ResponseWriter
// and a gzip.Writer to provide gzip compression for the response.
type gzipResponseWriter struct {
	Writer *gzip.Writer
	http.ResponseWriter
}

// ----------------------------------------------------
// Write method
// ----------------------------------------------------
// Write writes the given byte slice to the gzipResponseWriter's gzip.Writer.
// It returns the number of bytes written and any error that occurred during writing.
func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// GzipCompressHandler function
//
// GzipCompressHandler wraps an http.Handler with content compression middleware.
// It checks if the "Content-Encoding" header is present in the response. If it is,
// the function returns early without applying compression. Then, it checks if the
// "Accept-Encoding" header in the request contains the "gzip" encoding. If it does,
// the function sets the "Content-Encoding" header to "gzip" and creates a new gzip.Writer
// to compress the response. It defers the closing of the writer and wraps the original
// ResponseWriter with a gzipResponseWriter to intercept and compress the response's content.
// Finally, it calls ServeHTTP on the wrapped handler with the gzipResponseWriter and the original request.
func GzipCompressHandler(h http.Handler, o *GatewayOption) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ok := existContentEncoding(w); ok {
			h.ServeHTTP(w, r)
			return
		}

		if ok := withinAcceptEncoding(r, "gzip"); !ok {
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		writer := gzip.NewWriter(w)

		defer func() {
			if err := writer.Close(); err != nil {
				o.logger.Errorf("Error closing gzip writer: %v", err)
			}
		}()

		wo := &gzipResponseWriter{Writer: writer, ResponseWriter: w}
		h.ServeHTTP(wo, r)
	})
}

// BrotliCompressHandler function
//
// BrotliCompressHandler wraps an http.Handler with content compression middleware.
// It checks if the "Content-Encoding" header is present in the response. If it is,
// the function returns early without applying compression. Then, it checks if the
// "Accept-Encoding" header in the request contains the "br" encoding. If it does,
// the function sets the "Content-Encoding" header to "br" and creates a new brotli.Writer
// to compress the response. It defers the closing of the writer and wraps the original
// ResponseWriter with a brotliResponseWriter to intercept and compress the response's content.
// Finally, it calls ServeHTTP on the wrapped handler with the brotliResponseWriter and the original request.
func BrotliCompressHandler(h http.Handler, o *GatewayOption) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ok := existContentEncoding(w); ok {
			h.ServeHTTP(w, r)
			return
		}

		if ok := withinAcceptEncoding(r, "br"); !ok {
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "br")
		writer := brotli.NewWriter(w)

		defer func() {
			if err := writer.Close(); err != nil {
				o.logger.Errorf("Error closing brotli writer: %v", err)
			}
		}()

		wo := &brotliResponseWriter{Writer: writer, ResponseWriter: w}
		h.ServeHTTP(wo, r)
	})
}

func DeflateCompressHandler(h http.Handler, o *GatewayOption) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ok := existContentEncoding(w); ok {
			h.ServeHTTP(w, r)
			return
		}

		if ok := withinAcceptEncoding(r, "deflate"); !ok {
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "deflate")
		writer, err := flate.NewWriter(w, flate.BestCompression)

		if err != nil {
			o.logger.Errorf("Error creating flate writer: %v", err)
		}

		defer func() {
			if err := writer.Close(); err != nil {
				o.logger.Errorf("Error closing flate writer: %v", err)
			}
		}()

		wo := &deflateResponseWriter{Writer: writer, ResponseWriter: w}
		h.ServeHTTP(wo, r)
	})
}

// brotliResponseWriter is a type that wraps an http.ResponseWriter and a brotli.Writer
// to provide brotli compression for the response.
type brotliResponseWriter struct {
	Writer *brotli.Writer
	http.ResponseWriter
}

// ----------------------------------------------------
// Write method
// ----------------------------------------------------
// Write writes the given byte slice to the brotliResponseWriter's brotli.Writer.
// It returns the number of bytes written and any error that occurred during writing.
func (w *brotliResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

type deflateResponseWriter struct {
	Writer *flate.Writer
	http.ResponseWriter
}

func (w *deflateResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func chunkValues(value string) []string {
	var values []string

	for _, v := range strings.Split(value, ",") {
		v = strings.TrimSpace(v)

		if v != "" {
			values = append(values, v)
		}
	}

	return values
}

func withinAcceptEncoding(r *http.Request, needle string) bool {
	hit := map[string]struct{}{}

	for _, v := range chunkValues(r.Header.Get("Accept-Encoding")) {
		hit[v] = struct{}{}
	}

	_, ok := hit[needle]

	return ok
}

func existContentEncoding(w http.ResponseWriter) bool {
	value := strings.TrimSpace(w.Header().Get("Content-Encoding"))

	return value != ""
}
