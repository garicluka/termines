package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

func (a *app) log(data ...any) error {
	a.wg.Add(1)
	defer a.wg.Done()

	loggingFilePath, err := getLogsFilePath()
	if err != nil {
		return err
	}

	_, err = os.Stat(loggingFilePath)
	fileExists := !os.IsNotExist(err)
	if err != nil && fileExists {
		return err
	}

	file, err := os.OpenFile(loggingFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	_, fileName, line, ok := runtime.Caller(1)
	if !ok {
		fileName = "unknown_file"
		line = 0
	}
	cwd, err := os.Getwd()
	if err == nil {
		relPath, err := filepath.Rel(cwd, fileName)
		if err == nil {
			fileName = relPath
		}
	}

	_, err = file.WriteString(time.Now().Format("2006-01-02 15:04:05") + " " + fileName + ":" + strconv.Itoa(line) + "\n" + fmt.Sprint(data...) + "\n\n")
	return err
}

func getLogsFilePath() (string, error) {
	terminesDir, err := getTerminesDir()
	if err != nil {
		return "", err
	}

	logsFilePath := filepath.Join(terminesDir, "logs.txt")

	return logsFilePath, nil
}
