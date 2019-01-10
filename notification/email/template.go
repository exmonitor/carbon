package email

import (
	"fmt"
	"regexp"
)

const (
	emailName = "AlerTea"

	subjectEmailName      = "AlerTea"
	statusFailed          = `CRITICAL`
	statusResolved        = `Resolved`
	statusFailedColored   = `<font color="red">CRITICAL</font>`
	statusResolvedColored = `<font color="green">Resolved</font>`
)

const engEmailTemplate = `
<h2>%s</h2>
<p>
Your service '<b>%s - %s://%s%s</b>' has failed monitoring check:
</p>
 
<br>
<p>
Failure reason: %s
</p>
`

func (e *Email) emailBody() string {
	var body string
	status := stringStatusColored(e.failed)
	port := tryParsePort(e.serviceInfo.Metadata)
	body = fmt.Sprintf(engEmailTemplate, status, e.serviceInfo.Host, e.serviceInfo.ServiceTypeString(), e.serviceInfo.Target, port, e.failedMsg)

	return body
}

func (e *Email) emailSubject() string {
	//get string for bool value
	status := stringStatus(e.failed)
	// example '[AlerTea] CRITICAL: myhost - tcp:84.12.34.54'
	subject := fmt.Sprintf("[%s] %s: %s -  %s://%s", subjectEmailName, status, e.serviceInfo.Host, e.serviceInfo.ServiceTypeString(), e.serviceInfo.Target)

	return subject
}

func buildFromHeader(emailFrom string, name string) string {
	return fmt.Sprintf("%s <%s>", name, emailFrom)
}

func stringStatus(s bool) string {
	if s {
		return statusFailed
	} else {
		return statusResolved
	}
}

func stringStatusColored(s bool) string {
	if s {
		return statusFailedColored
	} else {
		return statusResolvedColored
	}
}

// try parse port number from service metadata
// in case of service without port, it will return empty string
func tryParsePort(metadata string) string {
	port := ""
	regExp := regexp.MustCompile(`"port": (\d*),`)

	result := regExp.FindStringSubmatch(metadata)
	if len(result) > 1 {
		port = fmt.Sprintf(":%s", result[1])
	}

	return port
}
