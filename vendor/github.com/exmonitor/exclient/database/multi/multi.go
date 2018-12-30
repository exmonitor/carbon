package multi

import (
	"database/sql"
	"fmt"

	"github.com/exmonitor/chronos"
	"github.com/exmonitor/exlogger"
	_ "github.com/go-sql-driver/mysql"
	"github.com/olivere/elastic"
	"github.com/pkg/errors"

	"context"
	"github.com/exmonitor/exclient/database"
)

const (
	sqlDriver = "mysql"

	esStatusIndex    = "service_status"
	esStatusDocName  = "service_status"
	esRangeQueryName = "my_range_query"
)

func DBDriverName() string {
	return "multi"
}

// config for multi db client with mariaDB and elastic search
type Config struct {
	// elastic search
	ElasticConnection string
	// maria db
	MariaConnection   string
	MariaUser         string
	MariaPassword     string
	MariaDatabaseName string

	Logger        *exlogger.Logger
	TimeProfiling bool
}

type Client struct {
	esClient  *elastic.Client
	sqlClient *sql.DB

	ctx           context.Context
	logger        *exlogger.Logger
	timeProfiling bool
	// implement client db interface
	database.ClientInterface
}

func New(conf Config) (*Client, error) {
	if conf.Logger == nil {
		return nil, errors.Wrapf(invalidConfigError, "conf.Logger must not be nil")
	}
	ctx := context.Background()

	// SQL
	sqlClient, err := createSqlClient(conf)
	if err != nil {
		return nil, err
	}

	// ELASTIC SEARCH
	esClient, err := createElasticsearchClient(conf, ctx)
	if err != nil {
		return nil, err
	}

	// init client
	newClient := &Client{
		esClient:  esClient,
		sqlClient: sqlClient,

		ctx:           ctx,
		logger:        conf.Logger,
		timeProfiling: conf.TimeProfiling,
	}
	return newClient, nil
}

func mysqlConnectionString(mariaConnection string, mariaUser string, mariaPassword string, mariaDatabaseName string) string {
	return fmt.Sprintf("%s:%s@%s/%s", mariaUser, mariaPassword, mariaConnection, mariaDatabaseName)
}

// close db connections
func (c *Client) Close() {
	c.sqlClient.Close()
	c.logger.Log("successfully closed sql connection")
	// there is no need for closing es client but just for consistency lets write it here
	c.logger.Log("successfully closed elasticsearch connection")
}

// initialise and check sql connection
func createSqlClient(conf Config) (*sql.DB, error) {
	// SQL
	t1 := chronos.New()
	// create sql connection string
	sqlConnectionString := mysqlConnectionString(conf.MariaConnection, conf.MariaUser, conf.MariaPassword, conf.MariaDatabaseName)
	// init sql connection
	sqlClient, err := sql.Open(sqlDriver, sqlConnectionString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create sql connection")
	}
	err = sqlClient.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "failed to ping sql connection")
	}
	t1.Finish()
	conf.Logger.Log("successfully connected to sql db %s", conf.MariaConnection)
	if conf.TimeProfiling {
		conf.Logger.LogDebug("TIME_PROFILING: created sql connection in %sms", t1.StringMilisec())
	}

	return sqlClient, nil
}

// initialise and check elasticsearch connection
func createElasticsearchClient(conf Config, ctx context.Context) (*elastic.Client, error) {
	t2 := chronos.New()
	// Create a client
	esClient, err := elastic.NewClient(elastic.SetURL(conf.ElasticConnection))

	if err != nil {
		return nil, errors.Wrap(err, "failed to create elasticsearch connection")
	}
	// check connection
	_, _, err = esClient.Ping(conf.ElasticConnection).Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ping elasticsearch")
	}
	// ensure status index is created
	_, err = esClient.CreateIndex(esStatusIndex).Do(ctx)
	if elastic.IsStatusCode(err, 400) {
		// all good, index already exists
		conf.Logger.LogDebug("Elasticsearch index '%s' already created, skipping", esStatusIndex)
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to create default index for elasticsearch")
	} else {
		conf.Logger.LogDebug("Elasticsearch index '%s' created", esStatusIndex)

	}

	t2.Finish()
	conf.Logger.Log("successfully connected to elasticsearch db %s", conf.ElasticConnection)
	if conf.TimeProfiling {
		conf.Logger.LogDebug("TIME_PROFILING: created elasticsearch connection in %sms, %ss", t2.StringMilisec(), t2.StringSecLong())
	}

	return esClient, nil
}
