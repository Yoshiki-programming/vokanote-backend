package Slog

import (
	"fmt"
	"log/slog"
	"os"
)

func DebugInfo(key any, value any) {
	if os.Getenv("GO_ENV") == "development" {
		slog.Info(fmt.Sprintf("ENV=%s", os.Getenv("GO_ENV")), key, value)
	}
}

func Error(err error) {
	slog.Error(fmt.Sprintf("ENV=%s", os.Getenv("GO_ENV")), "MSG", err)
}

func Warn(msg any) {
	slog.Warn(fmt.Sprintf("ENV=%s", os.Getenv("GO_ENV")), "MSG", msg)
}
