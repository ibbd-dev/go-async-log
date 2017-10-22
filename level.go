package asyncLog

import (
	"math/rand"
	"fmt"
)

// 日志优先级
type Priority int

const (
	LevelAll Priority = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelOff
)

var (
	// 日志等级
	levelTitle = map[Priority]string{
		LevelDebug: "[DEBUG]",
		LevelInfo:  "[INFO]",
		LevelWarn:  "[WARN]",
		LevelError: "[ERROR]",
		LevelFatal: "[FATAL]",
	}
)

// NewLevelLog 写入等级日志
// 级别高于logLevel才会被写入
func NewLevelLog(filename string, logLevel Priority) *LogFile {
	lf := NewLogFile(filename)
	lf.level = logLevel

	return lf
}

func (lf *LogFile) SetLevel(logLevel Priority) {
	lf.level = logLevel
}

func (lf *LogFile) Debug(format string, a ...interface{}) error {
	return lf.writeLevelMsg(LevelDebug, format, a...)
}

func (lf *LogFile) Info(format string, a ...interface{}) error {
	return lf.writeLevelMsg(LevelInfo, format, a...)
}

func (lf *LogFile) Warn(format string, a ...interface{}) error {
	return lf.writeLevelMsg(LevelWarn, format, a...)
}

func (lf *LogFile) Error(format string, a ...interface{}) error {
	return lf.writeLevelMsg(LevelError, format, a...)
}

func (lf *LogFile) Fatal(format string, a ...interface{}) error {
	return lf.writeLevelMsg(LevelFatal, format, a...)
}

func (lf *LogFile) writeLevelMsg(level Priority, format string, a ...interface{}) error {
	if lf.probability < 1.0 && rand.Float32() > lf.probability {
		// 按照概率写入
		return nil
	}

	if level >= lf.level {
		msg := fmt.Sprintf(format, a...)
		return lf.Write(levelTitle[level] + " " + msg)
	}

	return nil
}

