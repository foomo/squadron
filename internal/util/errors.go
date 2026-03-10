package util

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pterm/pterm"
)

func SprintError(err error) string {
	var ret strings.Builder

	prefix := "Error: "

	if pterm.PrintDebugMessages {
		return fmt.Sprintf("%+v", err) + "\n"
	}

	for {
		w := errors.Unwrap(err)
		if w == nil {
			ret.WriteString(prefix + err.Error() + "\n")
			break
		}

		if err.Error() != w.Error() {
			ret.WriteString(prefix + strings.TrimSuffix(err.Error(), ": "+w.Error()) + "\n")
			prefix = "↪ "
		}

		err = w
	}

	return strings.TrimSuffix(ret.String(), "\n")
}
