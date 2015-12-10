package log

import (
	"io/ioutil"
	"log"
	"os"
)

type Logger struct {
	traceLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errLogger   *log.Logger
	fatalLogger *log.Logger
}

func (l *Logger) Write(level string, format string, a ...interface{}) {
	switch level {
	case "INFO":
		l.infoLogger.Printf(format, a...)
	case "WARNING", "WARN":
		l.warnLogger.Printf(format, a...)
	case "ERROR", "ERR":
		l.errLogger.Printf(format, a...)
	case "FATAL":
		l.fatalLogger.Printf(format, a...)
		os.Exit(1)
	case "TRACE":
		l.traceLogger.Printf(format, a...)
	default:
		l.Write("FATAL", "Unknown logger level: %s\n", level)
	}
}

func New() *Logger {

	l := &Logger{}

	l.traceLogger = log.New(ioutil.Discard,
		"TRACE: \t",
		log.Ldate|log.Ltime)

	l.infoLogger = log.New(os.Stdout,
		"INFO: \t",
		log.Ldate|log.Ltime)

	l.warnLogger = log.New(os.Stdout,
		"WARNING: \t",
		log.Ldate|log.Ltime)

	l.errLogger = log.New(os.Stderr,
		"ERROR: \t",
		log.Ldate|log.Ltime)

	l.fatalLogger = log.New(os.Stderr,
		"FATAL: \t",
		log.Ldate|log.Ltime)

	return l
}
