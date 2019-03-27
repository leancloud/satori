package logging

import (
	"fmt"
	"log"
	"os"
)

const (
	LOG_FLAGS = log.Ldate | log.Lmicroseconds | log.Lshortfile
)

var (
	file *log.Logger
	std  = log.New(os.Stderr, "", LOG_FLAGS)
)

func init() {
	log.SetFlags(LOG_FLAGS)
	f, _ := os.OpenFile("/dev/null", os.O_WRONLY, 0222)
	log.SetOutput(f)
}

func SetOutputFilename(fn string) {
	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Can't open log file %s for writing: %s", fn, err)
	}
	log.SetOutput(f)
	file = log.New(f, "", LOG_FLAGS)
}

func Fatalln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	std.Output(2, s)
	file.Output(2, s)
	os.Exit(1)
}
