package logger

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
)

const (
	dbg = "[D]"
	inf = "[I]"
	err = "[E]"
	war = "[W]"

	pref = ""
	form = log.Ldate | log.Lmicroseconds

	lErr = 3
	lWar = 4
	lInf = 6
	lDbg = 7
)

// Log is a logging object
type Log struct {
	mu     sync.Mutex
	logger *log.Logger
	logbuf bytes.Buffer
	level  int
}

func NewStdErrLogger() *Log {
	return &Log{
		logger: log.New(os.Stderr, pref, form),
		level:  lDbg,
	}
}

func NewCaptureLogger() *Log {
	l := &Log{
		level: lDbg,
	}
	l.logger = log.New(&l.logbuf, pref, form)
	return l
}

func NewSilentLogger() *Log {
	return &Log{
		logger: log.New(ioutil.Discard, pref, form),
		level:  lDbg,
	}
}

func (l *Log) PrintLog() {
	fmt.Print(&l.logbuf)
}

func (l *Log) badLevel(lev int) bool {
	l.mu.Lock()
	b := l.level < lev
	l.mu.Unlock()
	return b
}

func (l *Log) SetLevel(lev int) {
	l.mu.Lock()
	l.level = lev
	l.mu.Unlock()
}

func (l *Log) prefix(level string, args ...interface{}) []interface{} {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	} else {
		bits := strings.Split(file, "/")
		if len(bits) > 2 {
			file = bits[len(bits)-2] + "/" + bits[len(bits)-1]
		}
	}
	a := []interface{}{level, fmt.Sprintf("(%s:%d)", file, line)}
	a = append(a, args...)
	return a
}

func (l *Log) Debug(args ...interface{}) {
	if l.badLevel(lDbg) {
		return
	}
	args = l.prefix(dbg, args...)
	l.logger.Println(args...)
}

func (l *Log) Debugf(format string, args ...interface{}) {
	if l.badLevel(lDbg) {
		return
	}
	args = []interface{}{fmt.Sprintf(format, args...)}
	args = l.prefix(dbg, args...)
	l.logger.Println(args...)
}

func (l *Log) Info(args ...interface{}) {
	if l.badLevel(lInf) {
		return
	}
	args = l.prefix(inf, args...)
	l.logger.Println(args...)
}

func (l *Log) Infof(format string, args ...interface{}) {
	if l.badLevel(lInf) {
		return
	}
	args = []interface{}{fmt.Sprintf(format, args...)}
	args = l.prefix(inf, args...)
	l.logger.Println(args...)
}

func (l *Log) Warning(args ...interface{}) {
	if l.badLevel(lWar) {
		return
	}
	args = l.prefix(war, args...)
	l.logger.Println(args...)
}

func (l *Log) Error(args ...interface{}) {
	if l.badLevel(lErr) {
		return
	}
	args = l.prefix(err, args...)
	l.logger.Println(args...)
}

func (l *Log) Errorf(format string, args ...interface{}) {
	if l.badLevel(lErr) {
		return
	}
	args = []interface{}{fmt.Sprintf(format, args...)}
	args = l.prefix(err, args...)
	l.logger.Println(args...)
}

func (l *Log) Fatal(args ...interface{}) {
	if l.badLevel(lErr) {
		return
	}
	args = l.prefix(err, args...)
	l.logger.Println(args...)
	os.Exit(255)
}
