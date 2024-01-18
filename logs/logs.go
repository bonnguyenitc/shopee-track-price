package logs

import (
	"os"

	logger "github.com/sirupsen/logrus"
)

func LogSetup() {
	file, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logger.Fatal(err)
	}

	defer file.Close()

	logger.SetOutput(file)
	logger.SetFormatter(&logger.JSONFormatter{})
}

func LogWarning(data logger.Fields, message string) {
	logger.SetLevel(logger.WarnLevel)
	logger.WithFields(data).Warn(message)
}
