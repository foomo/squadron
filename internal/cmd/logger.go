package cmd

import (
	"log/slog"
	"os"

	"github.com/pterm/pterm"
)

func init() {
	pterm.Info.Prefix.Text = "⎈"
	pterm.Info.Scope.Style = &pterm.ThemeDefault.DebugMessageStyle
	pterm.Debug.Prefix.Text = "⛏︎"
	pterm.Debug.Scope.Style = &pterm.ThemeDefault.DebugMessageStyle
	pterm.Fatal.Prefix.Text = "⛔︎"
	pterm.Fatal.Scope.Style = &pterm.ThemeDefault.DebugMessageStyle
	pterm.Error.Prefix.Text = "⛌"
	pterm.Error.Scope.Style = &pterm.ThemeDefault.DebugMessageStyle
	pterm.Warning.Prefix.Text = "⚠"
	pterm.Warning.Scope.Style = &pterm.ThemeDefault.DebugMessageStyle
	pterm.Success.Prefix.Text = "✓"
	pterm.Success.Scope.Style = &pterm.ThemeDefault.DebugMessageStyle

	if scope := os.Getenv("SQUADRON_SCOPE"); scope != "" {
		pterm.Info.Scope.Text = scope
		pterm.Debug.Scope.Text = scope
		pterm.Fatal.Scope.Text = scope
		pterm.Error.Scope.Text = scope
		pterm.Warning.Scope.Text = scope
		pterm.Success.Scope.Text = scope
	}
}

func NewLogger() *slog.Logger {
	return slog.New(NewPTermSlogHandler())
}
