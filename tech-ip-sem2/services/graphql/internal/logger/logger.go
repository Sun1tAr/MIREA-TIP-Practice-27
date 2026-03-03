package logger

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

type SpringFormatter struct {
	ServiceName string
	InstanceID  string
	PID         int
}

func (f *SpringFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")

	level := strings.ToUpper(entry.Level.String())
	switch level {
	case "WARNING":
		level = "WARN"
	case "ERROR":
		level = "ERROR"
	case "INFO":
		level = "INFO"
	case "DEBUG":
		level = "DEBUG"
	}

	pid := f.PID
	if pid == 0 {
		pid = os.Getpid()
	}

	thread := fmt.Sprintf("%s", f.ServiceName)
	if f.InstanceID != "" {
		thread = thread + "/" + f.InstanceID
	}

	var caller string
	if entry.HasCaller() {
		file := entry.Caller.File
		parts := strings.Split(file, "/")
		if len(parts) > 2 {
			file = strings.Join(parts[len(parts)-2:], ".")
		} else {
			file = strings.Join(parts, ".")
		}
		file = strings.TrimSuffix(file, ".go")
		caller = fmt.Sprintf("%s.%s", file, entry.Caller.Function)
	} else {
		caller = "?"
	}

	var levelColor int
	if f.useColors() {
		levelColor = f.getColor(entry.Level)
	}

	var logLine string
	if f.useColors() && levelColor != 0 {
		logLine = fmt.Sprintf("%s \x1b[%dm%5s\x1b[0m %d --- [%20s] %-30s : %s\n",
			timestamp, levelColor, level, pid, thread, caller, entry.Message)
	} else {
		logLine = fmt.Sprintf("%s %5s %d --- [%20s] %-30s : %s\n",
			timestamp, level, pid, thread, caller, entry.Message)
	}

	if len(entry.Data) > 0 {
		fields := make([]string, 0, len(entry.Data))
		for k, v := range entry.Data {
			fields = append(fields, fmt.Sprintf("%s=%v", k, v))
		}
		logLine += fmt.Sprintf("  %s\n", strings.Join(fields, ", "))
	}

	return []byte(logLine), nil
}

func (f *SpringFormatter) useColors() bool {
	return os.Getenv("TERM") != "dumb" && runtime.GOOS != "windows"
}

func (f *SpringFormatter) getColor(level logrus.Level) int {
	switch level {
	case logrus.DebugLevel:
		return 36
	case logrus.InfoLevel:
		return 32
	case logrus.WarnLevel:
		return 33
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return 31
	default:
		return 0
	}
}

func Init(serviceName string) *logrus.Logger {
	Logger = logrus.New()
	Logger.SetOutput(os.Stdout)
	Logger.SetReportCaller(true)
	Logger.SetFormatter(&SpringFormatter{
		ServiceName: serviceName,
		InstanceID:  os.Getenv("INSTANCE_ID"),
		PID:         os.Getpid(),
	})

	level := os.Getenv("LOG_LEVEL")
	if level != "" {
		lvl, err := logrus.ParseLevel(level)
		if err == nil {
			Logger.SetLevel(lvl)
		}
	} else {
		Logger.SetLevel(logrus.InfoLevel)
	}

	return Logger
}

func WithField(key string, value interface{}) *logrus.Entry {
	return Logger.WithField(key, value)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return Logger.WithFields(fields)
}

func WithError(err error) *logrus.Entry {
	return Logger.WithError(err)
}