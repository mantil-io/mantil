package log

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

const (
	levelInfo  = "info"
	levelDebug = "debug"
	levelError = "error"
)

type logLine struct {
	Level string    `json:"level"`
	Msg   string    `json:"msg"`
	Time  time.Time `json:"time"`
}

func Info(format string, v ...interface{}) {
	print(levelInfo, fmt.Sprintf(format, v...))
}

func Debug(format string, v ...interface{}) {
	print(levelDebug, fmt.Sprintf(format, v...))
}

func Error(err error) {
	Errorf(err.Error())
}

func Errorf(format string, v ...interface{}) {
	print(levelError, fmt.Sprintf(format, v...))
}

func print(level, msg string) {
	l := logLine{
		Level: level,
		Msg:   msg,
		Time:  time.Now(),
	}
	out, err := json.Marshal(l)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println(string(out))
}
