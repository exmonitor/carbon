package main

import (
	"github.com/pkg/errors"

	"github.com/exmonitor/firefly/database/dummydb"
	"github.com/exmonitor/firefly/database/multi"
	"github.com/exmonitor/firefly/database"
)


type DBConfig struct {
	DBDriver string
	// elastic search
	ElasticConnection string
	// maria db
	MariaConnection string
	MariaUser       string
	MariaPassword   string
}


func GetDBClient(conf DBConfig) (database.ClientInterface, error) {
	switch conf.DBDriver {
	case dummydb.DBDriverName():
		// dummydb has no errors on init
		c := dummydb.GetClient(dummydb.Config{})
		return c, nil
	case multi.DBDriverName():
		config := multi.Config{
			ElasticConnection: conf.ElasticConnection,
			MariaConnection:   conf.MariaConnection,
			MariaUser:         conf.MariaUser,
			MariaPassword:     conf.MariaPassword,
		}
		return multi.New(config)
	default:
		return nil, errors.Wrap(invalidDBDriver, conf.DBDriver)
	}
}
