package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()
