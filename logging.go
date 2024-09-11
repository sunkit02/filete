package main

import (
	"io"
	"log"
)

var (
	Trace   *log.Logger
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func initializeLoggers(out io.Writer) {

	flags := log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile

	Trace = log.New(out, "TRACE: ", flags)
	Debug = log.New(out, "DEBUG: ", flags)
	Info = log.New(out, "INFO: ", flags)
	Warning = log.New(out, "WARNING: ", flags)
	Error = log.New(out, "ERROR: ", flags)
}
