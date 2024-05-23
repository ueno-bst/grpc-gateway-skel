package runtime

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type CommonLogFormatter struct{}

func (f *CommonLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("02/Jan/2006:15:04:05 -0700")

	log := fmt.Sprintf(
		"%s - - [%s] \"%s %s %s\" %d %d %.3f %s \"%s\"\n",
		entry.Data["remote_addr"],
		timestamp,
		entry.Data["method"],
		entry.Data["url"],
		entry.Data["proto"],
		entry.Data["status_code"],
		entry.Data["size"],
		entry.Data["duration"],
		entry.Data["referer"],
		entry.Data["user_agent"],
	)

	return []byte(log), nil
}

func CommonLogHandler(h http.Handler, o *GatewayOption) http.Handler {
	o.log.SetFormatter(&CommonLogFormatter{})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		userAgent := r.UserAgent()

		if userAgent == "" {
			userAgent = "-"
		}

		referer := r.Referer()

		if referer == "" {
			referer = "-"
		}

		entry := o.log.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
			"method":      r.Method,
			"url":         r.URL.String(),
			"proto":       r.Proto,
			"referer":     referer,
			"user_agent":  userAgent,
		})

		lrw := &logResponseWriter{w, http.StatusOK, 0}
		h.ServeHTTP(lrw, r)

		entry = entry.WithFields(logrus.Fields{
			"status_code": lrw.statusCode,
			"size":        lrw.size,
			"duration":    time.Since(start).Seconds(),
		})

		entry.Info("request processed")
	})
}

type logResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (w *logResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *logResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}
