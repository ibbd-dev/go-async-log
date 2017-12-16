package asyncLog

import (
	"testing"
	"time"
)

func TestLevelWrite(t *testing.T) {
	infoFile := NewLevelLog("/tmp/test-info.log", LevelInfo)
	infoFile.Debug("hello %s", "world")
	infoFile.Info("hello %d", 123)
	infoFile.Error("hello %x", &t)

	time.Sleep(time.Second * 2)
}
