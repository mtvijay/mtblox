package mtlog

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"sync"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// A Level is a logging priority. Higher levels are more important.
//
// Note that Level satisfies the Option interface, so any Level can be passed to
// New to override the default logging priority.
type LogLevel int32

type LogConfig struct {
	LogToStdout bool     `json:"LogToStdout"`
	Level       LogLevel `json:"Level"`
	Path        string   `json:"Path"`
	File        string   `json:"File"`
	MaxBackups  int      `json:"MaxBackups"`
	MaxSize     int      `json:"MaxSize"`
	MaxAge      int      `json:"MaxAge"`
	Compress    bool     `json:"Compress"`
}

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel LogLevel = iota
	// InfoLevel is the default logging priority.
	InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel
	// PanicLevel logs a message, then panics.
	PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel
)

var LocalZLog Logger

func InitLogging(cfg LogConfig) {
	LocalZLog = NewLogfmtLogger(&cfg)
}

/*
 * If the input Writer is empty, create a file and write to that file
 * Interface with logging SaaS service
 * script to take level
 */
func NewLogfmtLogger(cfg *LogConfig) Logger {
	var l *log.Logger
	var level LogLevel
	flag := log.LstdFlags | log.Lmicroseconds
	if cfg == nil || cfg.LogToStdout == true || cfg.Path == "" || cfg.File == "" {
		l = log.New(os.Stdout, "", flag)
		level = InfoLevel
	} else {
		filename := path.Join(cfg.Path, cfg.File)
		lj := &lumberjack.Logger{Filename: filename, MaxSize: cfg.MaxSize,
			MaxBackups: cfg.MaxBackups, MaxAge: cfg.MaxAge, Compress: cfg.Compress}
		l = log.New(lj, "", flag)
		level = cfg.Level
	}
	return &logfmtLogger{
		lh:  l,
		lvl: level,
	}
}

// String returns a lower-case ASCII representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case PanicLevel:
		return "PANIC"
	case FatalLevel:
		return "FATAL"
	default:
		return fmt.Sprintf("Level(%d)", l)
	}
}

// A Logger enables leveled, structured logging. All methods are safe
// for concurrent use
type Logger interface {
	// Check the minimum enabled log level
	GetLevel() LogLevel
	// Change the level of this logger, as well as all of its ancestors
	// and decendants. This makes it easy to change the log level at runtime
	// without restarting the application
	SetLevel(LogLevel)

	// Create a child logger, and hopefully add some contex to that logger
	With(keyvals ...interface{}) Logger

	// Log a message at a given level. Messages include any context
	// that's accumulated on the logger, as well any fields added at the
	// log site
	Log(lvl LogLevel, keyvals ...interface{})
	Debug(keyvals ...interface{})
	Info(keyvals ...interface{})
	Warn(keyvals ...interface{})
	Error(keyvals ...interface{})
	Panic(keyvals ...interface{})
	Fatal(keyvals ...interface{})

	Tracef(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

func SetDefaultLogger(s Logger) {
	LocalZLog = s
}

type logfmtLogger struct {
	sync.Mutex
	lvl LogLevel
	lh  *log.Logger
}

// FIXME:
// this function needs more work, if with keyvals cannot set level to info
func (mtlog *logfmtLogger) With(keyvals ...interface{}) Logger {
	return &logfmtLogger{
		lvl: InfoLevel,
	}
}

func (mtlog *logfmtLogger) Log(lvl LogLevel, keyvals ...interface{}) {
	switch lvl {
	case PanicLevel:
		mtlog.Panic(keyvals...)
	case FatalLevel:
		mtlog.Fatal(keyvals...)
	default:
		mtlog.log(lvl, keyvals...)
	}
}

func (mtlog *logfmtLogger) Debug(keyvals ...interface{}) {
	mtlog.log(DebugLevel, keyvals...)
}

func (mtlog *logfmtLogger) Info(keyvals ...interface{}) {
	mtlog.log(InfoLevel, keyvals...)
}

func (mtlog *logfmtLogger) Warn(keyvals ...interface{}) {
	mtlog.log(WarnLevel, keyvals...)
}

func (mtlog *logfmtLogger) Error(keyvals ...interface{}) {
	mtlog.log(ErrorLevel, keyvals...)
}

func (mtlog *logfmtLogger) Panic(keyvals ...interface{}) {
	mtlog.log(PanicLevel, keyvals...)
	panic("Panic: Service going down")
}

func (mtlog *logfmtLogger) Fatal(keyvals ...interface{}) {
	mtlog.log(FatalLevel, keyvals...)
	os.Exit(1)
}

func (mtlog *logfmtLogger) GetLevel() LogLevel {
	mtlog.Lock()
	defer mtlog.Unlock()
	return mtlog.lvl
}

func (mtlog *logfmtLogger) SetLevel(lvl LogLevel) {
	mtlog.Lock()
	defer mtlog.Unlock()
	mtlog.lvl = lvl
}

func (mtlog *logfmtLogger) log(lvl LogLevel, keyvals ...interface{}) {
	if !(lvl >= mtlog.GetLevel()) {
		return
	}
	// prefix the keyvals with the level and we are all set!
	n := len(keyvals) + 2
	kvs := make([]interface{}, 0, n)
	kvs = append(append(kvs, "", lvl.String()), keyvals...)
	mtlog.lh.Printf("%v", kvs)
}

func (mtlog *logfmtLogger) Tracef(format string, args ...interface{}) {
	mtlog.logf(DebugLevel, format, args...)
}

func (mtlog *logfmtLogger) Infof(format string, args ...interface{}) {
	mtlog.logf(InfoLevel, format, args...)
}

func (mtlog *logfmtLogger) Warnf(format string, args ...interface{}) {
	mtlog.logf(WarnLevel, format, args...)
}

func (mtlog *logfmtLogger) Errorf(format string, args ...interface{}) {
	mtlog.logf(ErrorLevel, format, args...)
}

func (mtlog *logfmtLogger) logf(lvl LogLevel, format string, args ...interface{}) {
	if !(lvl >= mtlog.GetLevel()) {
		return
	}
	// prefix the keyvals with the level and we are all set!
	n := len(args) + 1
	kvs := make([]interface{}, 0, n)
	kvs = append(append(kvs, lvl.String()), args...)
	mtlog.lh.Printf("%s "+format, kvs...)
}

func Debug(keyvals ...interface{}) {
	if LocalZLog != nil {
		LocalZLog.Debug(keyvals)
	}
}

func Info(keyvals ...interface{}) {
	if LocalZLog != nil {
		LocalZLog.Info(keyvals)
	}
}

func Warn(keyvals ...interface{}) {
	if LocalZLog != nil {
		LocalZLog.Warn(keyvals)
	}
}

func Error(keyvals ...interface{}) {
	if LocalZLog != nil {
		LocalZLog.Error(keyvals)
	}
}

func Panic(keyvals ...interface{}) {
	if LocalZLog != nil {
		LocalZLog.Panic(keyvals)
	}
}

func Fatal(keyvals ...interface{}) {
	if LocalZLog != nil {
		LocalZLog.Fatal(keyvals)
	}
}

func Tracef(format string, args ...interface{}) {
	if LocalZLog != nil {
		LocalZLog.Tracef(format, args...)
	}
}

func Infof(format string, args ...interface{}) {
	if LocalZLog != nil {
		LocalZLog.Infof(format, args...)
	}
}

func Errorf(format string, args ...interface{}) {
	if LocalZLog != nil {
		LocalZLog.Errorf(format, args...)
	}
}

func Warnf(format string, args ...interface{}) {
	if LocalZLog != nil {
		LocalZLog.Warnf(format, args...)
	}
}

func SetLevel(lvl LogLevel) {
	if LocalZLog != nil {
		LocalZLog.SetLevel(lvl)
	}
}

func GetLevel() LogLevel {
	var lvl LogLevel = 0
	if LocalZLog != nil {
		lvl = LocalZLog.GetLevel()
	}
	return lvl
}

func JsonStringify(structStr interface{}) string {
	jsonStr, _ := json.Marshal(structStr)
	return string(jsonStr)
}
