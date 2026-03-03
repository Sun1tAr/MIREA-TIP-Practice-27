package logger

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// Logger - глобальный экземпляр логгера
var Logger *logrus.Logger

// SpringFormatter - кастомный форматтер в стиле Spring Boot
type SpringFormatter struct {
	ServiceName string
	InstanceID  string
	PID         int
}

// Format реализует logrus.Formatter
func (f *SpringFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Время: 2022-10-20 12:40:11.311
	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")

	// Уровень: INFO, WARN, ERROR (ровно 5 символов для выравнивания)
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

	// PID процесса
	pid := f.PID
	if pid == 0 {
		pid = os.Getpid()
	}

	// Имя потока/треда (в Go используем горутину)
	thread := fmt.Sprintf("%s", f.ServiceName)
	if f.InstanceID != "" {
		thread = thread + "/" + f.InstanceID
	}

	// Определяем caller (файл и метод)
	var caller string
	if entry.HasCaller() {
		// o.s.b.d.f.s.MyApplication
		file := entry.Caller.File
		// Берём только последние 2 части пути
		parts := strings.Split(file, "/")
		if len(parts) > 2 {
			file = strings.Join(parts[len(parts)-2:], ".")
		} else {
			file = strings.Join(parts, ".")
		}
		// Убираем .go
		file = strings.TrimSuffix(file, ".go")
		caller = fmt.Sprintf("%s.%s", file, entry.Caller.Function)
	} else {
		caller = "?"
	}

	// Формируем строку: 2022-10-20 12:40:11.311  INFO 16138 --- [           main] o.s.b.d.f.s.MyApplication                : Starting MyApplication...
	var levelColor int
	if f.useColors() {
		levelColor = f.getColor(entry.Level)
	}

	// Основной формат
	var logLine string
	if f.useColors() && levelColor != 0 {
		logLine = fmt.Sprintf("%s \x1b[%dm%5s\x1b[0m %d --- [%20s] %-30s : %s\n",
			timestamp, levelColor, level, pid, thread, caller, entry.Message)
	} else {
		logLine = fmt.Sprintf("%s %5s %d --- [%20s] %-30s : %s\n",
			timestamp, level, pid, thread, caller, entry.Message)
	}

	// Добавляем поля, если есть
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
	// Проверяем, поддерживает ли терминал цвета
	return os.Getenv("TERM") != "dumb" && runtime.GOOS != "windows"
}

func (f *SpringFormatter) getColor(level logrus.Level) int {
	switch level {
	case logrus.DebugLevel:
		return 36 // Cyan
	case logrus.InfoLevel:
		return 32 // Green
	case logrus.WarnLevel:
		return 33 // Yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return 31 // Red
	default:
		return 0
	}
}

// Init инициализирует структурированный логгер в стиле Spring
func Init(serviceName string) *logrus.Logger {
	Logger = logrus.New()

	// Настройка вывода
	Logger.SetOutput(os.Stdout)

	// Включаем caller (чтобы видеть, откуда лог)
	Logger.SetReportCaller(true)

	// Кастомный форматтер
	Logger.SetFormatter(&SpringFormatter{
		ServiceName: serviceName,
		InstanceID:  os.Getenv("INSTANCE_ID"),
		PID:         os.Getpid(),
	})

	// Уровень логирования
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

// WithField добавляет поле к логу
func WithField(key string, value interface{}) *logrus.Entry {
	return Logger.WithField(key, value)
}

// WithFields добавляет несколько полей
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Logger.WithFields(fields)
}

// WithError добавляет ошибку
func WithError(err error) *logrus.Entry {
	return Logger.WithError(err)
}
