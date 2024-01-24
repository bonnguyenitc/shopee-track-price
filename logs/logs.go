package logs

import (
	logger "github.com/sirupsen/logrus"
)

func LogWarning(data logger.Fields, message string) {
	logger.SetLevel(logger.WarnLevel)
	logger.WithFields(data).Warn(message)
}
