package logger

import (
	"io"
	"log"
	"os"
)

var (
	InfoLogger  *log.Logger
	WarnLogger  *log.Logger
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
)

const defaultLogFilePath = "./blockchain.log"

func Init(logFilePath ...string) {
	filePath := defaultLogFilePath
	if len(logFilePath) > 0 {
		filePath = logFilePath[0]
	}

	logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	multiWriter := io.MultiWriter(logFile, os.Stdout)

	InfoLogger = log.New(multiWriter, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	WarnLogger = log.New(multiWriter, "[WARN] ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	DebugLogger = log.New(logFile, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
}
