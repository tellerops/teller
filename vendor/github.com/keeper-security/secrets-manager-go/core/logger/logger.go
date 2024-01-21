package logger

import (
	"io"
	"log"
	"os"
	"runtime"
)

type LogLevel int

const (
	CriticalLevel LogLevel = iota + 1
	ErrorLevel
	WarningLevel
	NoticeLevel
	InfoLevel
	DebugLevel
)

type Logger struct {
	logInstance *log.Logger
	logLevel    LogLevel
}

var (
	klog Logger
)

func init() {
	klog = Logger{
		logInstance: log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds),
		logLevel:    WarningLevel,
	}
}

func (l *Logger) logLevelString() string {
	return logLevelString(l.logLevel)
}

func logLevelString(level LogLevel) string {
	if level >= CriticalLevel && level <= DebugLevel {
		logLevels := [...]string{
			"CRITICAL",
			"ERROR",
			"WARNING",
			"NOTICE",
			"INFO",
			"DEBUG",
		}
		return logLevels[level-1]
	}
	return ""
}

func logLevelTokenString(level LogLevel) string {
	return logLevelString(level) + ": "
}

func (l *Logger) SetLogLevel(level LogLevel) {
	if level >= CriticalLevel && level <= DebugLevel {
		l.logLevel = level
	}
}

func (l *Logger) Flags() int {
	return l.logInstance.Flags()
}

func (l *Logger) SetFlags(flag int) {
	l.logInstance.SetFlags(flag)
}

func (l *Logger) Prefix() string {
	return l.logInstance.Prefix()
}

func (l *Logger) SetPrefix(prefix string) {
	l.logInstance.SetPrefix(prefix)
}

func (l *Logger) Writer() io.Writer {
	return l.logInstance.Writer()
}

func (l *Logger) SetOutput(w io.Writer) {
	l.logInstance.SetOutput(w)
}

func (l *Logger) Log(level LogLevel, v ...interface{}) {
	l.logInternal(level, v...)
}

func (l *Logger) logInternal(level LogLevel, v ...interface{}) {
	if level <= l.logLevel && level >= CriticalLevel {
		token := logLevelTokenString(level)
		all := append([]interface{}{token}, v...)
		l.logInstance.Println(all...)
	}
}

// Note! Built-in Fatal != regular log message with status FATAL
func (l *Logger) Fatal(v ...interface{}) {
	l.logInstance.Fatal(v...)
}
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logInstance.Fatalf(format, v...)
}
func (l *Logger) Fatalln(v ...interface{}) {
	l.logInstance.Fatalln(v...)
}

func (l *Logger) Panic(v ...interface{}) {
	l.logInstance.Panic(v...)
}
func (l *Logger) Panicf(format string, v ...interface{}) {
	l.logInstance.Panicf(format, v...)
}
func (l *Logger) Panicln(v ...interface{}) {
	l.logInstance.Panicln(v...)
}

func (l *Logger) Print(v ...interface{}) {
	l.logInstance.Print(v...)
}
func (l *Logger) Printf(format string, v ...interface{}) {
	l.logInstance.Printf(format, v...)
}
func (l *Logger) Println(v ...interface{}) {
	l.logInstance.Println(v...)
}

func (l *Logger) PrintStack(level LogLevel, message string) {
	if message == "" {
		message = "Stack info"
	}
	message += "\n"
	l.Log(ErrorLevel, message+Stack())
}

func Stack() string {
	buf := make([]byte, 1000000)
	runtime.Stack(buf, false)
	return string(buf)
}

// public interface
func Fatal(v ...interface{}) {
	klog.logInstance.Fatal(v...)
}
func Fatalf(format string, v ...interface{}) {
	klog.logInstance.Fatalf(format, v...)
}
func Fatalln(v ...interface{}) {
	klog.logInstance.Fatalln(v...)
}

func Panic(v ...interface{}) {
	klog.logInstance.Panic(v...)
}
func Panicf(format string, v ...interface{}) {
	klog.logInstance.Panicf(format, v...)
}
func Panicln(v ...interface{}) {
	klog.logInstance.Panicln(v...)
}

func Print(v ...interface{}) {
	klog.logInstance.Print(v...)
}
func Printf(format string, v ...interface{}) {
	klog.logInstance.Printf(format, v...)
}
func Println(v ...interface{}) {
	klog.logInstance.Println(v...)
}

func Log(level LogLevel, v ...interface{}) {
	klog.Log(level, v...)
}

func Critical(v ...interface{}) {
	klog.Log(CriticalLevel, v...)
}
func Error(v ...interface{}) {
	klog.Log(ErrorLevel, v...)
}
func Warning(v ...interface{}) {
	klog.Log(WarningLevel, v...)
}
func Notice(v ...interface{}) {
	klog.Log(NoticeLevel, v...)
}
func Info(v ...interface{}) {
	klog.Log(InfoLevel, v...)
}
func Debug(v ...interface{}) {
	klog.Log(DebugLevel, v...)
}

func SetLogLevel(level LogLevel) {
	klog.SetLogLevel(level)
}

func Flags() int {
	return klog.Flags()
}

func SetFlags(flag int) {
	klog.SetFlags(flag)
}

func Prefix() string {
	return klog.Prefix()
}

func SetPrefix(prefix string) {
	klog.SetPrefix(prefix)
}

func Writer() io.Writer {
	return klog.Writer()
}

func SetOutput(w io.Writer) {
	klog.SetOutput(w)
}
