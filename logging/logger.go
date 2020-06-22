package logging

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

const kCallerSkip int = 2

type Logger struct {
	level Level
	pool  *bufferPool
	w     io.Writer
}

func New() *Logger {
	logger := &Logger{
		level: TraceLevel,
		pool:  newBufferPool(),
		w:     os.Stdout,
	}
	return logger
}

func (logger *Logger) SetOutput(w io.Writer) {
	logger.w = w
}

func (logger *Logger) SetLevel(level Level) {
	atomic.StoreInt32((*int32)(&logger.level), int32(level))
}

func (logger *Logger) GetLevel() Level {
	return Level(atomic.LoadInt32((*int32)(&logger.level)))
}

func (logger *Logger) IsLevelDisabled(level Level) bool {
	return logger.GetLevel() > level
}

func (logger *Logger) output(l Level, s string) {
	now := time.Now() // get this early.
	_, file, line, ok := runtime.Caller(kCallerSkip)
	if !ok {
		file = "???"
		line = 1
	}
	buf := logger.pool.get()

	buf.formatHeader(now, l, file, line)

	buf.appendString(s)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		buf.appendByte('\n')
	}

	if l > ErrorLevel {
		// FIXME: Sync()
	}

	logger.w.Write(buf.bytes())
	logger.pool.put(buf)
}

func (logger *Logger) log(l Level, s string) {
	if logger.level.Enabled(l) {
		logger.output(l, s)
	}
}

func (logger *Logger) Tracef(format string, v ...interface{}) {
	logger.log(TraceLevel, fmt.Sprintf(format, v...))
}

func (logger *Logger) Trace(v ...interface{}) {
	logger.log(TraceLevel, fmt.Sprint(v...))
}

func (logger *Logger) Debugf(format string, v ...interface{}) {
	logger.log(DebugLevel, fmt.Sprintf(format, v...))
}

func (logger *Logger) Debug(v ...interface{}) {
	logger.log(DebugLevel, fmt.Sprint(v...))
}

func (logger *Logger) Infof(format string, v ...interface{}) {
	logger.log(InfoLevel, fmt.Sprintf(format, v...))
}

func (logger *Logger) Info(v ...interface{}) {
	logger.log(InfoLevel, fmt.Sprint(v...))
}

func (logger *Logger) Warningf(format string, v ...interface{}) {
	logger.log(WarningLevel, fmt.Sprintf(format, v...))
}

func (logger *Logger) Warning(v ...interface{}) {
	logger.log(WarningLevel, fmt.Sprint(v...))
}

func (logger *Logger) Errorf(format string, v ...interface{}) {
	logger.log(ErrorLevel, fmt.Sprintf(format, v...))
}

func (logger *Logger) Error(v ...interface{}) {
	logger.log(ErrorLevel, fmt.Sprint(v...))
}

func (logger *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	logger.log(PanicLevel, s)
	panic(s)
}

func (logger *Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	logger.log(PanicLevel, s)
	panic(s)
}

func (logger *Logger) Fatalf(format string, v ...interface{}) {
	logger.log(FatalLevel, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (logger *Logger) Fatal(v ...interface{}) {
	logger.log(FatalLevel, fmt.Sprint(v...))
	os.Exit(1)
}
