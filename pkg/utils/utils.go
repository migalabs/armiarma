package utils

import (
	"github.com/sirupsen/logrus"
)

var (
	ModuleName = "UTILS"
	log        = logrus.WithField(
		"module", ModuleName,
	)
)
