package util_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/foomo/squadron/internal/util"
	errorsx "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestSprintError(t *testing.T) {
	err := errors.New("test error")
	err = errorsx.Wrap(err, "1 wrap")
	err = errorsx.WithMessage(err, "2 with message")

	ret := util.SprintError(err)

	t.Log(ret)
	assert.Len(t, strings.Split(ret, "\n"), 3)
}
