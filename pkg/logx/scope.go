package logx

import (
	"github.com/sirupsen/logrus"
)

func Scope(logger logrus.FieldLogger, scope string) logrus.FieldLogger {
	return logger.WithField("scope", scope)
}

func Component(logger logrus.FieldLogger, component string) logrus.FieldLogger {
	return logger.WithField("component", component)
}
