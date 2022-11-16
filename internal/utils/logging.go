package internal

import (
	"os"

	"go.elastic.co/ecszap"
	"go.uber.org/zap"
)

// SetUpLogger sets up and returns an ecs logger.
func SetUpLogger() zap.Logger {
	encoderConfig := ecszap.NewDefaultEncoderConfig()
	core := ecszap.NewCore(encoderConfig, os.Stdout, zap.DebugLevel)
	logger := zap.New(core, zap.AddCaller())
	logger = logger.Named("NamespaceLabelLogger")

	return *logger
}
