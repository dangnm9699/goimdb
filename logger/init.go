package logger

import (
	"os"
)

var F *os.File

func init() {
	f, err := os.OpenFile("logger.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	F = f
}

func WriteLog(msg string) {
	if _, err := F.WriteString(msg); err != nil {
		panic(err)
	}
}