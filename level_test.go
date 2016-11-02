package asyncLog

import (
	"testing"
	"time"
)

func TestLevelWrite(t *testing.T) {
	infoFile := NewLevelLog("/tmp/test-info.log", LevelInfo)
	infoFile.Debug("hello world")
	infoFile.Info("hello world")
	infoFile.Error("hello world")

	time.Sleep(time.Second * 2)
}
