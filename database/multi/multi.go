package multi

import (
	"github.com/exmonitor/firefly/database"
)


func DBDriverName() string {
	return "multi"
}

// config for multi db client with mariaDB and elastic search
type Config struct {
	// elastic search
	ElasticConnection string
	// maria db
	MariaConnection string
	MariaUser       string
	MariaPassword   string
}

type Client struct {
	// TODO

	// implement client db interface
	database.ClientInterface
}

func New(conf Config) (*Client, error) {
	// TODO
	return nil, nil
}
