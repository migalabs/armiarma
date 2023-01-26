package db

import (
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var (
	ModuleName = "MSG-DB"
	log        = logrus.WithField(
		"module", ModuleName,
	)
	PSQLHost   = "localhost"
	PSQLPort   = "5432"
	PSQLUser   = "armiarma"
	PSQLPasswd = "localhost"
	PSQLDbName = "localhost"
)

// type struct {

// }
