package logger

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/apperr"
)

type CloseLogger func() error

func InitializeLogger(logFile string) (*slog.Logger, CloseLogger, error) {
	debugHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:       slog.LevelDebug,
		ReplaceAttr: replaceAttr,
	})

	file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open log file: %w", err)
	}

	bufferedWriter := bufio.NewWriterSize(file, 8192)
	closeLogger := func() error {
		err := bufferedWriter.Flush()
		if err != nil {
			return err
		}
		err = file.Close()
		if err != nil {
			return err
		}
		return nil
	}

	infoHandler := slog.NewJSONHandler(bufferedWriter, &slog.HandlerOptions{
		Level:       slog.LevelInfo,
		ReplaceAttr: replaceAttr,
	})

	logger := slog.New(slog.NewMultiHandler(
		debugHandler,
		infoHandler,
	))

	return logger, closeLogger, nil
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "error" {
		err, ok := a.Value.Any().(error)
		if !ok {
			return a
		}
		errorAttrs := []slog.Attr{
			slog.String("message", err.Error()),
		}
		errorAttrs = append(errorAttrs, apperr.Attrs(err)...)
		return slog.GroupAttrs("error", errorAttrs...)
	}
	return a
}
