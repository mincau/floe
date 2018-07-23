package config

import (
	golog "log"
	"sync"
)

var sg sync.Once

func setupLog() {
	sg.Do(func() {
		log = &tLog{}
	})
}

type tLog struct{}

func (l *tLog) Info(args ...interface{}) {
	golog.Println("INFO", args)
}
func (l *tLog) Debug(args ...interface{}) {
	golog.Println("DEBUG", args)
}
func (l *tLog) Error(args ...interface{}) {
	golog.Println("Error", args)
}
func (l *tLog) Warning(args ...interface{}) {
	golog.Println("WARNING", args)
}
func (l *tLog) Debugf(format string, args ...interface{}) {
	golog.Printf("DEBUG - "+format+"\n", args...)
}
