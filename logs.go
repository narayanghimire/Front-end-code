package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
)

//ENV LOGGER_ENABLE "" to disable

type Logger interface {
	Log(w http.ResponseWriter, r *http.Request, e *HttpError, t time.Duration)
}

var logger Logger

func init() {
	if os.Getenv("LOGGER_ENABLE") != "" {
		logger = NewMuxLogger(NewFileLogger(), NewTtyLogger())
	}
}

type TtyLogger struct { //uses default logger
}

func (f *TtyLogger) Log(w http.ResponseWriter, r *http.Request, e *HttpError, t time.Duration) {
	if e != nil {
		log.Printf("%s %s %s error %d %s headers request{%#v} response{%#v}\n", r.Method, r.URL, t, e.Code, e.Message, r.Header, w.Header())
	} else {
		log.Printf("%s %s %s headers request{%#v} response{%#v}\n", r.Method, r.URL, t, r.Header, w.Header())
	}
}

func NewTtyLogger() Logger {
	return &TtyLogger{}
}

const (
	INTERNAL_SERVER_ERROR_LOGS_FILE_PATH = "logs/internal-server-error.log"
	HTTP_REQUEST_LOGS_FILE_PATH          = "logs/http-request.log"
	BAD_REQUEST_LOGS_FILE_PATH           = "logs/bad-request.log"
)

type FileLogger struct {
	HttpRequestLogFile         *os.File
	InternalServerErrorLogFile *os.File
	BadRequestLogFile          *os.File

	HttpRequestLogger         *log.Logger
	InternalServerErrorLogger *log.Logger
	BadRequestLogger          *log.Logger
}

func (f *FileLogger) Log(w http.ResponseWriter, r *http.Request, e *HttpError, t time.Duration) {
	if e != nil {
		var l *log.Logger
		switch e.Code {
		case http.StatusInternalServerError:
			l = f.InternalServerErrorLogger
		case http.StatusBadRequest:
			l = f.BadRequestLogger
		default:
			l = f.HttpRequestLogger
		}
		l.Printf("%s %s %s error %d %s headers request{%#v} response{%#v}\n", r.Method, r.URL, t, e.Code, e.Message, r.Header, w.Header())
	} else {
		f.HttpRequestLogger.Printf("%s %s %s headers request{%#v} response{%#v}\n", r.Method, r.URL, t, r.Header, w.Header())
	}
}

func NewFileLogger() Logger {
	if _, err := os.Stat("logs"); os.IsNotExist(err) { // if logs folder does not exits then create it
		if err = os.Mkdir("logs", 0774); err != nil {
			fmt.Println("unable to create logs", err)
			return nil
		}
	}

	if _, err := os.Stat("files"); os.IsNotExist(err) { //  if files folder does not exits then create it
		if err = os.Mkdir("files", 0770); err != nil {
			fmt.Println("unable to create files", err)
			return nil
		}
	}

	var files FileLogger
	var err error
	files.InternalServerErrorLogFile, err = os.OpenFile(INTERNAL_SERVER_ERROR_LOGS_FILE_PATH, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v\n", err)
		return nil
	}

	files.HttpRequestLogFile, err = os.OpenFile(HTTP_REQUEST_LOGS_FILE_PATH, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v\n", err)
		return nil
	}

	files.BadRequestLogFile, err = os.OpenFile(BAD_REQUEST_LOGS_FILE_PATH, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v\n", err)
		return nil
	}

	files.InternalServerErrorLogger = log.New(files.InternalServerErrorLogFile, "", log.LstdFlags)
	files.HttpRequestLogger = log.New(files.HttpRequestLogFile, "", log.LstdFlags)
	files.BadRequestLogger = log.New(files.BadRequestLogFile, "", log.LstdFlags)
	return &files
}

type MuxLogger struct { //multiplex multiple loggers
	l []Logger
}

func NewMuxLogger(x ...Logger) Logger {
	m := &MuxLogger{}
	m.l = x
	return m
}

func (m *MuxLogger) Log(w http.ResponseWriter, r *http.Request, e *HttpError, t time.Duration) {
	for _, f := range m.l {
		if f != nil {
			f.Log(w, r, e, t)
		}
	}
}

func LogHandler(a AppHandler) AppHandler {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) *HttpError {
		t := time.Now()
		e := a(w, r, p)
		if logger != nil {
			logger.Log(w, r, e, time.Since(t))
		}
		return e
	}
}
