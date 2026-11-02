package main

import (
	"fmt"
	"log"
	"os"
)

type processLogger struct {
	file      *os.File
	logger    *log.Logger
	printToIO bool
}

func (p processLogger) LogInfo(message string) {
	prefix := "INFO: "
	p.logger.Println(prefix + message)
	if p.printToIO {
		fmt.Println(prefix + message)
	}
}

func (p processLogger) LogError(message string) {
	prefix := "ERROR: "
	p.logger.Println(prefix + message)
	if p.printToIO {
		fmt.Println(prefix + message)
	}
}

func (p processLogger) EndSpace() {
	p.logger.Println("---")
}

func (p processLogger) Close() {
	p.file.Close()
}

type logError struct {
	errMessage string
	logMessage string
}

func (l *logError) Error() string {
	return l.errMessage
}

func (l *logError) LogError() string {
	return l.logMessage
}

type logInfo struct {
	logMessage string
}

type logResponse struct {
	logError logError
	logInfo  logInfo
}

type logResponder struct {
	logResponses []logResponse
}

func (l *logResponder) AddInfo(msg string) {
	l.logResponses = append(l.logResponses, logResponse{
		logInfo: logInfo{
			logMessage: msg,
		},
	})
}

func (l *logResponder) AddError(msg string) {
	l.logResponses = append(l.logResponses, logResponse{
		logError: logError{
			logMessage: msg,
		},
	})
}

func (l *logResponder) AddResponses(ll *logResponder) {
	l.logResponses = append(l.logResponses, ll.logResponses...)
}

func (p processLogger) Log(l *logResponder) {
	for _, response := range l.logResponses {
		if response.logError.logMessage != "" {
			p.LogError(response.logError.logMessage)
		} else {
			p.LogInfo(response.logInfo.logMessage)
		}
	}
}

func initProcessLogger(filePath string) processLogger {
	prefix := ""

	processLogFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("error intitating process log: %s\n", err.Error())
	}

	logger := log.New(processLogFile, prefix, log.Ldate|log.Ltime)

	return processLogger{
		file:   processLogFile,
		logger: logger,
	}

}
