package henchman

import (
	logrus "gopkg.in/Sirupsen/logrus.v0"
	"os"
)

var jsonLog = logrus.New()

func InitLog() {
	jsonLog.Level = logrus.DebugLevel
	jsonLog.Formatter = new(logrus.JSONFormatter)

	// NOTE: hardcoded for now
	f, _ := os.Create("DummyLogsForHenchman")
	jsonLog.Out = f
}

// wrapper for debug
func Debug(fields map[string]interface{}, msg string) {
	if DebugFlag {
		if fields == nil {
			jsonLog.Debug(msg)
		} else {
			jsonLog.WithFields(fields).Debug(msg)
		}
	}
}

// wrapper for Info
func Info(fields map[string]interface{}, msg string) {
	if fields == nil {
		jsonLog.Info(msg)
	} else {
		jsonLog.WithFields(fields).Info(msg)
	}
}

// wrapper for Fatal
func Fatal(fields map[string]interface{}, msg string) {
	if fields == nil {
		jsonLog.Fatal(msg)
		logrus.Fatal(msg)
	} else {
		jsonLog.WithFields(fields).Fatal(msg)
		logrus.WithFields(fields).Fatal(msg)
	}
}

// wrapper for Error
func Error(fields map[string]interface{}, msg string) {
	if fields == nil {
		jsonLog.Error(msg)
		logrus.Error(msg)
	} else {
		jsonLog.WithFields(fields).Error(msg)
		logrus.WithFields(fields).Error(msg)
	}
}

func Warn(fields map[string]interface{}, msg string) {
	if fields == nil {
		jsonLog.Warn(msg)
		logrus.Warn(msg)
	} else {
		jsonLog.WithFields(fields).Warn(msg)
		logrus.WithFields(fields).Warn(msg)
	}
}
