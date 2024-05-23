package runtime

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
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

	o.logger = logrus.New()

	if o.logs.access != nil {
		if file, err := os.OpenFile(*o.logs.access, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err != nil {
			return fmt.Errorf("failed to open access log file %s for output: %s", *o.logs.access, err)
		} else {
			aws = append(aws, file)
		}
	}

	if o.logs.error != nil {
		if file, err := os.Create(*o.logs.error); err != nil {
			return fmt.Errorf("failed to open error log file %s for output: %s", *o.logs.error, err)
		} else {
			ews = append(ews, file)
		}
	}

	if !o.silent {
		aws = append(aws, os.Stdout)
		ews = append(ews, os.Stderr)
	}

	o.logger.Hooks.Add(&outputHook{writer: io.MultiWriter(aws...)})
	o.logger.Hooks.Add(&errorHook{writer: io.MultiWriter(ews...)})

	return nil
}

type errorHook struct {
	writer io.Writer
}

func (h *errorHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.DebugLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

func (h *errorHook) Fire(entry *logrus.Entry) error {
	entry.Logger.Out = h.writer
	return nil
}

type outputHook struct {
	writer io.Writer
}

func (h *outputHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.InfoLevel,
	}
}

func (h *outputHook) Fire(entry *logrus.Entry) error {
	entry.Logger.Out = h.writer
	return nil
}
