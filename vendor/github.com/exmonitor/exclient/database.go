package exclient

import (
	"github.com/pkg/errors"

	"github.com/exmonitor/exclient/database"
	"github.com/exmonitor/exclient/database/dummydb"
	"github.com/exmonitor/exclient/database/multi"
	"github.com/exmonitor/exlogger"
)

type DBConfig struct {
	DBDriver string
	// elastic search
	ElasticConnection string
	// maria db
	MariaConnection   string
	MariaDatabaseName string
	MariaUser         string
	MariaPassword     string

	Logger        *exlogger.Logger
	TimeProfiling bool
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
			MariaDatabaseName: conf.MariaDatabaseName,
			MariaUser:         conf.MariaUser,
			MariaPassword:     conf.MariaPassword,

			Logger:        conf.Logger,
			TimeProfiling: conf.TimeProfiling,
		}
		return multi.New(config)
	default:
		return nil, errors.Wrap(invalidDBDriver, conf.DBDriver)
	}
}
