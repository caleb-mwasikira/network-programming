package utils

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

type Logger struct {
	filename string
	*log.Logger
}

var default_logger = &Logger{
	filename: "",
	Logger:   log.New(os.Stdout, log.Default().Prefix(), log.Lshortfile),
}

func CreateLogger(fname string) *Logger {
	log_dir := filepath.Join(ProjectPath, ".logs/")
	err := os.MkdirAll(log_dir, 0700)
	if err != nil {
		log.Printf("failed to create log directory; %v! Switching to default logger\n", err)
		return default_logger
	}

	// TODO: change file flag to os.O_APPEND in production
	log_fpath := filepath.Join(log_dir, fname)
	file, err := os.OpenFile(log_fpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		log.Printf("failed to open log file; %v! Switching to default logger\n", err)
		return default_logger
	}

	multi_wrt := io.MultiWriter(os.Stdout, file)
	return &Logger{
		filename: log_fpath,
		Logger:   log.New(multi_wrt, log.Default().Prefix(), log.Lshortfile),
	}
}
