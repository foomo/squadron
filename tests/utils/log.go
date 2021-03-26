package testutils

import (
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func Log() *logrus.Entry {
	testLogger, _ := test.NewNullLogger()
	return logrus.NewEntry(testLogger)
}
