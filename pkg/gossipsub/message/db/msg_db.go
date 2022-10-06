package db

import (
	"github.com/sirupsen/logrus"
	"database/sql"
	_ "github.com/lib/pq"

)

var (
	ModuleName = "MSG-DB"
	log        = logrus.WithField(
		"module", ModuleName,
	)
	PSQLHost = "localhost"
	PSQLPort = "5432"
	PSQLUser = "armiarma" 
	PSQLPasswd = "localhost"
	PSQLDbName = "localhost"
)


type struct {
	
}