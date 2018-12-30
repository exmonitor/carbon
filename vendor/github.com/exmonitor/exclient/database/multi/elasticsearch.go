package multi

import (
	"fmt"
	"time"

	"github.com/exmonitor/chronos"
	"github.com/exmonitor/exclient/database/spec/status"
	"github.com/olivere/elastic"
	"github.com/pkg/errors"
	"reflect"
)

// **************************************************
// ELASTIC SEARCH
///--------------------------------------------------
func (c *Client) ES_GetFailedServices(from time.Time, to time.Time, interval int) ([]*status.ServiceStatus, error) {
	var serviceStatusArray []*status.ServiceStatus
	t := chronos.New()

	// datetime range query
	timeRangeQuery := elastic.NewRangeQuery("@timestamp").Gte(from).Lt(to)
	// failedServices term query
	faildServiceQuery := elastic.NewTermQuery("result", false)

	// build whole search query
	searchQuery := elastic.NewBoolQuery().Must(faildServiceQuery).Filter(timeRangeQuery)

	// execute search querry
	// TODO use backoff retry
	searchResult, err := c.esClient.Search().Index(esStatusIndex).Query(searchQuery).Do(c.ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ES_GetFailedServices for int %d", interval)
	}

	if c.timeProfiling {
		c.logger.LogDebug("TIME_PROFILING: executed search query in ES_SaveServiceStatus in %sms", searchResult.TookInMillis)
	}

	// parse results into struct
	var ttyp status.ServiceStatus
	for i, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
		if s, ok := item.(status.ServiceStatus); ok {
			serviceStatusArray = append(serviceStatusArray, &s)
		} else {
			// TODO should we exit ??
			c.logger.LogError(nil, "failed to parse status.ServiceStatus num %d in ES_SaveServiceStatus", i)
		}
	}
	c.logger.LogDebug("fetched %d FailedServices from db", len(serviceStatusArray))

	t.Finish()
	if c.timeProfiling {
		c.logger.LogDebug("TIME_PROFILING: executed ES_SaveServiceStatus in %sms", t.StringMilisec())
	}

	return serviceStatusArray, nil
}

func (c *Client) ES_SaveServiceStatus(s *status.ServiceStatus) error {
	t := chronos.New()

	// insert data to elasticsearch db
	_, err := c.esClient.Index().Index(esStatusIndex).Type(esStatusDocName).BodyJson(s).Do(c.ctx)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to save service status for id %d", s.Id))
	}

	t.Finish()
	if c.timeProfiling {
		c.logger.LogDebug("TIME_PROFILING: executed ES_SaveServiceStatus:ID:%d in %sms", s.Id, t.StringMilisec())
	}
	return nil
}
