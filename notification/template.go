package notification

import (
	"github.com/exmonitor/exclient/database/spec/service"
)

func SMSTemplate(failed bool, service *service.Service) string {
	var s string
	if failed {
		s = "TODO / NOT IMPLEMENTED"
	} else {
		s = "TODO / NOT IMPLEMENTED"
	}

	return s
}

func CallTemplate(failed bool, service *service.Service) string {
	var s string
	if failed {
		s = "TODO / NOT IMPLEMENTED"
	} else {
		s = "TODO / NOT IMPLEMENTED"
	}

	return s
}
