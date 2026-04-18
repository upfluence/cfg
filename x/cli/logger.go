package cli

import (
	"io"
	"strings"

	"github.com/upfluence/errors"
	"github.com/upfluence/log"
	"github.com/upfluence/log/record"
	"github.com/upfluence/log/sink"
	"github.com/upfluence/log/sink/leveled"
	"github.com/upfluence/log/sink/writer"
)

type logLevel record.Level

var logLevels = map[string]record.Level{
	"debug":   record.Debug,
	"info":    record.Info,
	"notice":  record.Notice,
	"warning": record.Warning,
	"error":   record.Error,
}

func (ll *logLevel) Parse(v string) error {
	lvl, ok := logLevels[strings.ToLower(v)]

	if !ok {
		return errors.Newf("unknown log level %q", v)
	}

	*ll = logLevel(lvl)

	return nil
}

type bareFormatter struct{}

func (bareFormatter) Format(w io.Writer, r record.Record) error {
	r.WriteFormatted(w)

	return nil
}

type routingSink struct {
	stdout sink.Sink
	stderr sink.Sink
}

func (s *routingSink) Log(r record.Record) error {
	if r.Level() >= record.Error {
		return errors.Wrap(s.stderr.Log(r), "log to stderr")
	}

	return errors.Wrap(s.stdout.Log(r), "log to stdout")
}

func newLogger(stdout, stderr io.Writer, lvl record.Level) log.Logger {
	var f bareFormatter

	return log.NewLogger(
		log.WithSink(
			leveled.NewSink(
				lvl,
				&routingSink{
					stdout: writer.NewSink(f, stdout),
					stderr: writer.NewSink(f, stderr),
				},
			),
		),
	)
}
