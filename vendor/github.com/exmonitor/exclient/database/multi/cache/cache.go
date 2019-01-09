package cache

import (
	"time"

	"github.com/pkg/errors"
	"fmt"
)

type CacheSystemConfig struct {
	Enabled bool
	TTL     time.Duration
}

func New(conf CacheSystemConfig) (*CacheSystem, error) {
	if conf.TTL == 0 {
		return nil, errors.Wrap(invalidConfigError, "conf.TTL must not be zero")
	}

	if conf.TTL < time.Minute {
		fmt.Printf("WARNING: cache TTL under 1m doesnt make sense much sense as lowest interval is 30s and in each run cache will be expired.\n")
	}

	sql_getUserNotificationSettings := &SQL_GetUsersNotificationSetting{
		Cache: map[int]SQL_GetUsersNotificationSetting_Record{},
	}
	sql_getServices := &SQL_GetServices{
		Cache: map[int]SQL_GetServices_Record{},
	}
	sql_getServiceDetail := &SQL_GetServiceDetails{
		Cache: map[int]SQL_GetServiceDetails_Record{},
	}

	sqlCache := &SQLCache{
		GetUsersNotificationSettings: sql_getUserNotificationSettings,
		GetServices:                  sql_getServices,
		GetServiceDetails:            sql_getServiceDetail,
	}

	newCacheSystem := &CacheSystem{
		Enabled: conf.Enabled,
		TTL:     conf.TTL,
		SQL:     sqlCache,
	}
	return newCacheSystem, nil
}

type CacheSystem struct {
	Enabled bool
	TTL     time.Duration

	// caches
	SQL *SQLCache
}

type SQLCache struct {
	// SQL_GetUsersNotificationSettings
	GetUsersNotificationSettings *SQL_GetUsersNotificationSetting
	// SQL_GetServices
	GetServices *SQL_GetServices
	// SQL_GetServiceDetails
	GetServiceDetails *SQL_GetServiceDetails
}
