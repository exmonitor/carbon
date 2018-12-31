package exlogger

import (
	"io"
	"log"
	"os"

	"github.com/pkg/errors"

	"fmt"
)

type Config struct {
	Debug        bool
	LogToFile    bool
	LogErrorFile string
	LogFile      string
}

func New(conf Config) (*Logger, error) {
	var logOutput, logOutputError io.Writer
	if conf.LogToFile {
		f1, err := os.Open(conf.LogFile)
		if err != nil {
			return nil, errors.Wrap(err, "failed to log file")
		}
		f2, err := os.Open(conf.LogErrorFile)
		if err != nil {
			return nil, errors.Wrap(err, "failed to log error file")

		}
		logOutput = f1
		logOutputError = f2
	} else {
		logOutput = os.Stdout
		logOutputError = os.Stderr
	}

	newLogger := &Logger{
		debug:       conf.Debug,
		logger:      log.New(logOutput, "", log.LstdFlags),
		loggerError: log.New(logOutputError, "", log.LstdFlags),
	}

	return newLogger, nil
}

type Logger struct {
	debug       bool
	logger      *log.Logger
	loggerError *log.Logger

	f1 *os.File
	f2 *os.File
}

func (l *Logger) CloseLogs() {
	l.f1.Close()
	l.f2.Close()
	fmt.Printf("Closed files for logs.\n")
}

func (l *Logger) Log(msg string, vals ...interface{}) {
	l.logger.Printf("INFO | "+msg, vals...)
}
func (l *Logger) LogDebug(msg string, vals ...interface{}) {
	if l.debug {
		l.logger.Printf("DEBUG | "+msg, vals...)
	}
}

func (l *Logger) LogError(err error, msg string, vals ...interface{}) {
	l.loggerError.Printf(fmt.Sprintf("ERROR | %s|%s", msg, err), vals...)
}
