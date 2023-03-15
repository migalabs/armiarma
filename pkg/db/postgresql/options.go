package postgresql

import (
	"github.com/pkg/errors"
)

type DBOption func(*DBClient) error 

func InitializeTables(init bool) DBOption {
	return func (dbCli *DBClient) error {
		// initialize all the tables
		if init {
		    err := dbCli.initTables()
			if err != nil {
				return errors.Wrap(err, "unable to initialize the SQL tables at "+dbCli.loginStr)
			}
		}
		return nil
	}
}

func WithConnectionEventsPersist(persist bool) DBOption {
	return func (dbCli *DBClient) error {
		dbCli.persistConnEvents = persist
		return nil
	}
}

