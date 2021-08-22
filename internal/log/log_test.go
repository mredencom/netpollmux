package log

import (
	"log"
	"os"
	"testing"
)

func TestPrefix(t *testing.T) {
	var prefix = "log"
	SetPrefix(prefix)
	if prefix != GetPrefix() {
		t.Errorf("error %s != %s", prefix, GetPrefix())
	}
}

func TestLevel(t *testing.T) {
	var level = DebugLevel
	SetLevel(level)
	SetOut(os.Stdout)
	if level != GetLevel() {
		t.Errorf("error %d != %d", level, GetLevel())
	}
}

func TestMicroseconds(t *testing.T) {
	SetMicroseconds(true)
	if logger.microseconds != log.Lmicroseconds {
		t.Errorf("error %d != %d", logger.microseconds, log.Lmicroseconds)
	}
	SetMicroseconds(false)
	if logger.microseconds != 0 {
		t.Errorf("error %d != %d", logger.microseconds, 0)
	}
}

func TestDebug(t *testing.T) {
	Debug(1024, "HelloWorld", true)
	Debugf("%d %s %t", 1024, "HelloWorld", true)
	Debugln(1024, "HelloWorld", true)
}

func TestTrace(t *testing.T) {
	Trace(1024, "HelloWorld", true)
	Tracef("%d %s %t", 1024, "HelloWorld", true)
	Traceln(1024, "HelloWorld", true)
}

func TestAll(t *testing.T) {
	All(1024, "HelloWorld", true)
	Allf("%d %s %t", 1024, "HelloWorld", true)
	Allln(1024, "HelloWorld", true)
}

func TestInfo(t *testing.T) {
	Info(1024, "HelloWorld", true)
	Infof("%d %s %t", 1024, "HelloWorld", true)
	Infoln(1024, "HelloWorld", true)
}

func TestNotice(t *testing.T) {
	Notice(1024, "HelloWorld", true)
	Noticef("%d %s %t", 1024, "HelloWorld", true)
	Noticeln(1024, "HelloWorld", true)
}

func TestWarn(t *testing.T) {
	Warn(1024, "HelloWorld", true)
	Warnf("%d %s %t", 1024, "HelloWorld", true)
	Warnln(1024, "HelloWorld", true)
}

func TestError(t *testing.T) {
	Error(1024, "HelloWorld", true)
	Errorf("%d %s %t", 1024, "HelloWorld", true)
	Errorln(1024, "HelloWorld", true)
}

func TestPanic(t *testing.T) {
	Panic(1024, "HelloWorld", true)
	Panicf("%d %s %t", 1024, "HelloWorld", true)
	Panicln(1024, "HelloWorld", true)
}

func TestFatal(t *testing.T) {
	Fatal(1024, "HelloWorld", true)
	Fatalf("%d %s %t", 1024, "HelloWorld", true)
	Fatalln(1024, "HelloWorld", true)
}
