package logs

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// NewFileLogger creates and returns a Logrus logger that writes JSON-formatted logs
// to the specified file, within the given directory. If the directory doesn't exist,
// it will be created. The caller can also specify a log level (e.g., logrus.InfoLevel).
func Filelogger(dirPath, fileName string, level logrus.Level) (*logrus.Logger, error) {
	// Ensure the directory structure exists.
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory '%s': %v", dirPath, err)
	}

	// Open/create the log file in append mode.
	fullPath := dirPath + "/" + fileName
	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {

		return nil, fmt.Errorf("failed to open log file '%s': %v", fullPath, err)
	}

	// Create and configure the logger.
	logger := logrus.New()
	logger.Out = file
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
	logger.SetLevel(level)

	return logger, nil
}
