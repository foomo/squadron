package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pterm/pterm"
)

type PTermSlogHandler struct {
	attrs []slog.Attr
}

// Enabled returns true if the given level is enabled.
func (s *PTermSlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	switch level {
	case slog.LevelDebug:
		return pterm.PrintDebugMessages
	default:
		return true
	}
}

// Handle handles the given record.
func (s *PTermSlogHandler) Handle(ctx context.Context, record slog.Record) error {
	level := record.Level
	message := record.Message

	// Convert slog Attrs to a map.
	keyValsMap := make(map[string]any)

	record.Attrs(func(attr slog.Attr) bool {
		keyValsMap[attr.Key] = attr.Value
		return true
	})

	for _, attr := range s.attrs {
		keyValsMap[attr.Key] = attr.Value
	}

	args := pterm.DefaultLogger.ArgsFromMap(keyValsMap)

	// Wrapping args inside another slice to match [][]LoggerArgument
	argsWrapped := [][]pterm.LoggerArgument{args}

	for _, arg := range argsWrapped {
		for _, attr := range arg {
			message += " " + attr.Key + ": " + fmt.Sprintf("%v", attr.Value)
		}
	}

	switch level {
	case slog.LevelDebug:
		pterm.Debug.Println(message)
	case slog.LevelInfo:
		pterm.Info.Println(message)
	case slog.LevelWarn:
		pterm.Warning.Println(message)
	case slog.LevelError:
		pterm.Error.Println(message)
	default:
		pterm.Info.Println(message)
	}

	return nil
}

// WithAttrs returns a new handler with the given attributes.
func (s *PTermSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newS := *s
	newS.attrs = attrs
	return &newS
}

// WithGroup is not yet supported.
func (s *PTermSlogHandler) WithGroup(name string) slog.Handler {
	// Grouping is not yet supported by pterm.
	return s
}

// NewPTermSlogHandler returns a new logging handler that can be intrgrated with log/slog.
func NewPTermSlogHandler() *PTermSlogHandler {
	return &PTermSlogHandler{}
}
