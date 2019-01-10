package multi

import (
	"fmt"
	"time"

	"encoding/json"
	"github.com/exmonitor/chronos"
	"github.com/exmonitor/exclient/database/spec/status"
	"github.com/olivere/elastic"
	"github.com/pkg/errors"
	"reflect"
	"io"
)

// **************************************************
// ELASTIC SEARCH
///--------------------------------------------------
func (c *Client) ES_GetFailedServices(from time.Time, to time.Time, interval int) ([]*status.ServiceStatus, error) {

	failedServiceQuery := elastic.NewTermQuery("result", false)
	// match interval
	intervalQuery := elastic.NewTermQuery("interval", interval)

	serviceStatusArray, err := c.ES_GetServicesStatus(from, to, failedServiceQuery, intervalQuery)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute ES_GetFailedServices")
	}

	return serviceStatusArray, nil
}

func (c *Client) ES_GetServicesStatus(from time.Time, to time.Time, elasticQuery ...elastic.Query) ([]*status.ServiceStatus, error) {
	var serviceStatusArray []*status.ServiceStatus
	t := chronos.New()

	// datetime range query
	timeRangeFilter := elastic.NewRangeQuery("@timestamp").Gte(from).Lt(to)

	// build whole search query
	searchQuery := elastic.NewBoolQuery().Must(elasticQuery...).Filter(timeRangeFilter)

	// execute search querry
	// aggregated
	// TODO use backoff retry
	scrollService := c.esClient.Scroll(esStatusIndex).Size(scrollWindowSize).Query(searchQuery)
	for {
		result , err := scrollService.Do(c.ctx)
		if err == io.EOF {
			// end of scroll window, lets exit
			break
		}
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get ES_GetServicesStatus")
		}

		// parse results into struct
		var ttyp status.ServiceStatus
		for i, item := range result.Each(reflect.TypeOf(ttyp)) {
			if s, ok := item.(status.ServiceStatus); ok {
				serviceStatusArray = append(serviceStatusArray, &s)
			} else {
				// TODO should we exit ??
				c.logger.LogError(nil, "failed to parse status.ServiceStatus num %d in ES_GetServicesStatus", i)
			}
		}
	}


	t.Finish()
	if c.timeProfiling {
		c.logger.LogDebug("TIME_PROFILING: executed ES_GetServicesStatus in %sms, fetched %d results", t.StringMilisec(), len(serviceStatusArray))
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

func (c *Client) ES_DeleteServicesStatus(from time.Time, to time.Time) error {
	t := chronos.New()

	timeQuery := elastic.NewRangeQuery("@timestamp").Gte(from).Lt(to)

	// delete data from elasticsearch db
	o, err := c.esClient.DeleteByQuery().Index(esStatusIndex).Type(esStatusDocName).Query(timeQuery).Do(c.ctx)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to delete service status for range %s - %s", from.String(), to.String()))
	}
	t.Finish()
	if c.timeProfiling {
		c.logger.LogDebug("TIME_PROFILING: executed ES_DeleteServicesStatus in %sms, deleted %d records", t.StringMilisec(), o.Total)
	}
	return nil
}

func (c *Client) ES_GetAggregatedServiceStatusByID(from time.Time, to time.Time, serviceID int) (*status.AgregatedServiceStatus, error) {
	var aggregatedServiceStatus *status.AgregatedServiceStatus
	t := chronos.New()
	// search specific serviceID
	termQuery := elastic.NewTermQuery("service_id", serviceID)
	// datetime range query
	timeRangeFilter := elastic.NewRangeQuery("@timestamp_to").Gte(from).Lt(to)
	// compile query
	searchQuery := elastic.NewBoolQuery().Must(termQuery).Filter(timeRangeFilter)

	// execute search querry
	// TODO use backoff retry , Sort("@timestamp_to", true)
	searchResult, err := c.esClient.Search().Index(esAggregatedStatusIndex).Query(searchQuery).Sort("@timestamp_to", false).Size(1).Do(c.ctx)
	if err != nil {
		c.logger.LogDebug("cannot sort by timestamp_to, possible no record for serviceID")
		searchResult, err = c.esClient.Search().Index(esAggregatedStatusIndex).Query(searchQuery).Size(1).Do(c.ctx)
		if err != nil {

			return nil, errors.Wrapf(err, "failed to get ES_GetAggregatedServicesStatus")
		}
	}
	// parse result
	if searchResult.Hits.TotalHits > 0 {
		for _, rawItem := range searchResult.Hits.Hits {
			var item status.AgregatedServiceStatus
			err := json.Unmarshal(*rawItem.Source, &item)
			if err != nil {
				c.logger.LogError(err, "failed to parse status.AgregatedServiceStatus in ES_GetAggregatedServicesStatus")
			}
			// save elastic internal ID
			item.Id = rawItem.Id

			aggregatedServiceStatus = &item
		}
	}

	t.Finish()
	if c.timeProfiling {
		c.logger.LogDebug("TIME_PROFILING: executed ES_GetAggregatedServicesStatus in %sms", t.StringMilisec())
	}

	return aggregatedServiceStatus, nil
}

// insert new or update existing aggregated_service_status
func (c *Client) ES_SaveAggregatedServiceStatus(s *status.AgregatedServiceStatus) error {
	var err error
	t := chronos.New()
	// insert data to elasticsearch db, if the record already exists, use update function
	if s.Id != "" {
		// update existing document
		_, err = c.esClient.Update().Index(esAggregatedStatusIndex).Type(esAggregatedStatusDocName).Id(s.Id).Doc(s).Do(c.ctx)
	} else {
		// insert new document
		_, err = c.esClient.Index().Index(esAggregatedStatusIndex).Type(esAggregatedStatusDocName).BodyJson(s).Do(c.ctx)
	}

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to save aggregated_service status for serviceID %d", s.ServiceID))
	}

	t.Finish()
	if c.timeProfiling {
		if s.Id != "" {
			c.logger.LogDebug("TIME_PROFILING: executed ES_SaveAggregatedServiceStatus operation 'Update' in %sms", t.StringMilisec())
		} else {
			c.logger.LogDebug("TIME_PROFILING: executed ES_SaveAggregatedServiceStatus operation 'Insert' in %sms", t.StringMilisec())
		}
	}
	return nil

}
