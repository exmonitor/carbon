package email

import "fmt"

const (
	emailName = "AlerTea"

	subjectEmailName = "AlerTea"
	statusFailed     = "CRITICAL"
	statusResolved   = "Resolved"
)

const engEmailTemplate = `
<h3>%s: %s - %s:%s</h3>

<p>
Failure reason: %s
</p>
`

func (e *Email) emailBody() string {
	var body string
	status := stringStatus(e.failed)

	body = fmt.Sprintf(engEmailTemplate, status, e.serviceInfo.Host, e.serviceInfo.ServiceTypeString(), e.serviceInfo.Target, "TODO")

	return body
}

func (e *Email) emailSubject() string {
	//get string for bool value
	status := stringStatus(e.failed)
	// example '[AlerTea] CRITICAL: myhost - tcp:84.12.34.54'
	subject := fmt.Sprintf("[%s] %s: %s -  %s:%s", subjectEmailName, status, e.serviceInfo.Host, e.serviceInfo.ServiceTypeString(), e.serviceInfo.Target)

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
