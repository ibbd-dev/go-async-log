package asyncLog

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

func (lf *LogFile) Debug(msg string) error {
	return lf.writeLevelMsg(msg, LevelDebug)
}

func (lf *LogFile) Info(msg string) error {
	return lf.writeLevelMsg(msg, LevelInfo)
}

func (lf *LogFile) Warn(msg string) error {
	return lf.writeLevelMsg(msg, LevelWarn)
}

func (lf *LogFile) Error(msg string) error {
	return lf.writeLevelMsg(msg, LevelError)
}

func (lf *LogFile) Fatal(msg string) error {
	return lf.writeLevelMsg(msg, LevelFatal)
}

func (lf *LogFile) writeLevelMsg(msg string, level Priority) error {
	if level >= lf.level {
		return lf.Write(levelTitle[level] + " " + msg)
	}

	return nil
}
