package logging

import (
	"io"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	defaultLogger = New()
)

type Logger interface {
	Panic(fmt string, a ...interface{})
	Fatal(fmt string, a ...interface{})
	Error(fmt string, a ...interface{})
	Warn(fmt string, a ...interface{})
	Info(fmt string, a ...interface{})
	Debug(fmt string, a ...interface{})
	Trace(fmt string, a ...interface{})

	WithFields(map[string]interface{}) Logger
	WithField(string, interface{}) Logger
	WithError(err error) Logger

	GetLevel() string
	SetLevel(string)

	SetOutputFormat(format string)
	SetCallerReporter()
}

type StandardLogger struct {
	logger *logrus.Logger
	fields map[string]interface{}
}

// New returns a new application logger.
func New() *StandardLogger {
	return &StandardLogger{
		logger: logrus.New(),
	}
}

// GetRoot returns the root logger application.
func GetRoot() *StandardLogger {
	return defaultLogger
}

// SetOutput sets the underlying logrus output.
func (l *StandardLogger) SetOutput(w io.Writer) {
	l.logger.SetOutput(w)
}

// SetOutputFormat change the logger format.
// available formats: text/json
func (l *StandardLogger) SetOutputFormat(format string) {
	var formatter logrus.Formatter

	switch strings.ToLower(format) {
	case "text":
		formatter = &logrus.TextFormatter{
			FullTimestamp:          true,
			DisableLevelTruncation: true,
			PadLevelText:           true,
			QuoteEmptyFields:       true,
		}
	case "json":
		formatter = &logrus.JSONFormatter{
			PrettyPrint: false,
		}
	default:
		return // using default logger format
	}

	l.logger.SetFormatter(formatter)
}

// SetCallerReporter add calling method as a field.
func (l *StandardLogger) SetCallerReporter() {
	l.logger.SetReportCaller(true)
}

// WithFields create new logger instance with the given default fields.
//
// Example:
// logger :=log.WithFields(map[string]interface{}{"test": 2222})
// logger.Info("Teller message") -> `INFO[0000] Teller message  test=2222`
func (l *StandardLogger) WithFields(fields map[string]interface{}) Logger {
	cp := *l
	cp.fields = make(map[string]interface{})
	for k, v := range l.fields {
		cp.fields[k] = v
	}
	for k, v := range fields {
		cp.fields[k] = v
	}
	return &cp
}

// WithField create new logger instance with singe field.
func (l *StandardLogger) WithField(name string, value interface{}) Logger {
	return l.WithFields(map[string]interface{}{name: value})
}

// WithError add an error as single field to the logger.
func (l *StandardLogger) WithError(err error) Logger {
	return l.WithField("error", err)
}

// GetLevel return the current logging level.
func (l *StandardLogger) GetLevel() string {
	return l.logger.GetLevel().String()
}

// SetLevel sets the logger level.
func (l *StandardLogger) SetLevel(level string) {

	switch level {
	case "panic'":
		l.logger.SetLevel(logrus.PanicLevel)
	case "fatal":
		l.logger.SetLevel(logrus.FatalLevel)
	case "error":
		l.logger.SetLevel(logrus.ErrorLevel)
	case "warn", "warning":
		l.logger.SetLevel(logrus.WarnLevel)
	case "info":
		l.logger.SetLevel(logrus.InfoLevel)
	case "debug":
		l.logger.SetLevel(logrus.DebugLevel)
	case "trace":
		l.logger.SetLevel(logrus.TraceLevel)
	case "null", "none":
		l.logger.SetOutput(io.Discard)
	default:
		l.Warn("unknown log level %v", level)
		l.logger.SetLevel(logrus.ErrorLevel)
	}
}

func (l *StandardLogger) Panic(fmt string, a ...interface{}) {
	l.logger.WithFields(l.getFields()).Panicf(fmt, a...)
}

func (l *StandardLogger) Fatal(fmt string, a ...interface{}) {
	l.logger.WithFields(l.getFields()).Fatalf(fmt, a...)
}

func (l *StandardLogger) Error(fmt string, a ...interface{}) {
	l.logger.WithFields(l.getFields()).Errorf(fmt, a...)
}

func (l *StandardLogger) Warn(fmt string, a ...interface{}) {
	l.logger.WithFields(l.getFields()).Errorf(fmt, a...)
}

func (l *StandardLogger) Info(fmt string, a ...interface{}) {
	l.logger.WithFields(l.getFields()).Infof(fmt, a...)
}

func (l *StandardLogger) Debug(fmt string, a ...interface{}) {
	l.logger.WithFields(l.getFields()).Debugf(fmt, a...)
}

func (l *StandardLogger) Trace(fmt string, a ...interface{}) {
	l.logger.WithFields(l.getFields()).Tracef(fmt, a...)
}

func (l *StandardLogger) getFields() map[string]interface{} {
	return l.fields
}
