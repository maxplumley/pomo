package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/natefinch/lumberjack"
	log "github.com/sirupsen/logrus"
)

const logDir = "$HOME/.pomo/_logs"
const logFile = "pomo.log"

// verbosity flags
var (
	verbose = flag.Bool("v", false, "enable verbose output")
	debug   = flag.Bool("vv", false, "enable debug output")
)

func init() {
	flag.Parse()
}

func initLogger() error {
	// Determine log level based on flags
	var level log.Level = log.WarnLevel
	if *debug {
		level = log.DebugLevel
	} else if *verbose {
		level = log.InfoLevel
	}

	// Create log directory if it doesn't exist
	logDirPath := os.ExpandEnv(logDir)
	if err := os.MkdirAll(logDirPath, 0755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	lumberjackLogger := &lumberjack.Logger{
		// Log file abbsolute path, os agnostic
		Filename:   filepath.ToSlash(filepath.Join(logDirPath, logFile)),
		MaxSize:    5, // MB
		MaxBackups: 10,
		MaxAge:     30, // days
		Compress:   true,
	}

	writer := io.Writer(lumberjackLogger)

	logFormatter := new(log.TextFormatter)
	logFormatter.TimestampFormat = time.RFC3339
	logFormatter.FullTimestamp = true

	log.SetOutput(writer)
	log.SetLevel(level)
	log.SetFormatter(logFormatter)

	return nil
}
