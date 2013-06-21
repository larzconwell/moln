package main

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// Create a new basic error logger for the given file
func NewErrorLogger(file string) (*log.Logger, *os.File, error) {
	err := os.MkdirAll(filepath.Dir(file), os.ModeDir|os.ModePerm)
	if err != nil {
		return nil, nil, err
	}

	logFile, err := os.OpenFile(file, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, nil, err
	}

	return log.New(logFile, "", log.LstdFlags), logFile, nil
}

// NewLogFile creates or opens a log file using logPath as the log directory. A new log
// file is only create if there's no existing log files or if the newest log file is older
// than a week old
func NewLogFile(logPath string) (*os.File, error) {
	var logFile *os.File

	err := os.MkdirAll(logPath, os.ModeDir|os.ModePerm)
	if err != nil {
		return nil, err
	}

	logDir, err := os.OpenFile(logPath, os.O_RDONLY, os.ModeDir|os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer logDir.Close()

	logFiles, err := logDir.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	// If log files exists, get the most recent one
	if len(logFiles) > 0 {
		for _, filename := range logFiles {
			if filename == "errors" {
				continue
			}
			fileDate, err := time.Parse(time.RFC3339, filename)
			if err != nil {
				return nil, err
			}

			// If less than a week old then open it
			if time.Since(fileDate).Hours() <= 168 {
				logPath = filepath.Join(logPath, filename)
				logFile, err = os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND, os.ModePerm)
				if err != nil {
					return nil, err
				}
				break
			}
		}
	}

	// Create a new log file if non are in use
	if logFile == nil {
		logPath = filepath.Join(logPath, time.Now().Format(time.RFC3339))
		logFile, err = os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	return logFile, nil
}
