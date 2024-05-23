package runtime

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

func WithAccessLogOutput(path string) GatewayOptionFunc {
	return func(opt *GatewayOption) {
		opt.logs.access = &path
	}
}

func WithErrorOutput(path string) GatewayOptionFunc {
	return func(opt *GatewayOption) {
		opt.logs.error = &path
	}
}

func (o *GatewayOption) initLog() error {
	aws := make([]io.Writer, 0)
	ews := make([]io.Writer, 0)

	o.log = logrus.New()
	o.err = logrus.New()

	o.err.SetFormatter(&errorLogFormatter{})

	if o.logs.access != nil {
		if file, err := os.OpenFile(*o.logs.access, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err != nil {
			return fmt.Errorf("failed to open access log file %s for output: %s", *o.logs.access, err)
		} else {
			aws = append(aws, file)
		}
	}

	if o.logs.error != nil {
		if file, err := os.OpenFile(*o.logs.error, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err != nil {
			return fmt.Errorf("failed to open error log file %s for output: %s", *o.logs.error, err)
		} else {
			ews = append(ews, file)
		}
	}

	if !o.silent {
		aws = append(aws, os.Stdout)
		ews = append(ews, os.Stderr)
	}

	o.log.SetOutput(io.MultiWriter(aws...))
	o.err.SetOutput(io.MultiWriter(ews...))

	return nil
}

type errorLogFormatter struct{}

func (f *errorLogFormatter) Format(e *logrus.Entry) ([]byte, error) {
	log := fmt.Sprintf(
		"[%s] %s : %s\n",
		e.Time.Format("02/Jan/2006:15:04:05 -0700"),
		strings.ToTitle(e.Level.String()),
		strings.TrimSpace(e.Message),
	)

	return []byte(log), nil
}
