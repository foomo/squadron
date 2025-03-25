package util

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pterm/pterm"
)

func SprintError(err error) string {
	var ret string
	prefix := "Error: "
	if pterm.PrintDebugMessages {
		return fmt.Sprintf("%+v", err) + "\n"
	}

	for {
		w := errors.Unwrap(err)
		if w == nil {
			ret += prefix + err.Error() + "\n"
			break
		}
		if err.Error() != w.Error() {
			ret += prefix + strings.TrimSuffix(err.Error(), ": "+w.Error()) + "\n"
			prefix = "↪ "
		}
		err = w
	}

	return strings.TrimSuffix(ret, "\n")
}
