package email

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
)

const (
	emailName = "AlerTea"

	statusFailed   = `CRITICAL`
	statusResolved = `Resolved`
)

const emailTemplateCritical_ENG = `
<html>
 <body>
   <h2>
     <font color="red">
       CRITICAL
     </font>
   </h2> 
   <p>
     Your service has <b>failed</b> monitoring check.
   </p>
   <table>
     <tr>
       <td><label>Host:</label></td>
       <td><a rel="nofollow" style='text-decoration:none;'>{{ .Host }}</a></td>
     </tr>
     <tr>
       <td><label>DNS/IP:</label></td>
       <td><a rel="nofollow" style='text-decoration:none;'>{{ .Target }}</a></td>
     </tr>
     <tr>
       <td><label>Check type:</label></td>
       <td>{{ .ServiceType }}</td>
     </tr>
     <tr>
       <td><label>Port:</label></td>
       <td>{{ .Port }}</td>
     </tr>  
   </table>
   <br>
   <p>
     Failure reason: <a rel="nofollow" style='text-decoration:none;'> {{ .FailMessage }} </a>
   </p>
 </body>
</html>
`

const emailTemplateOK_ENG = `
<html>
 <body>
   <h2>
     <font color="green">
       Resolved
     </font>
   </h2>
 
   <p>
     Your service has <b>passed</b> monitoring check.
   </p>
   <table>
     <tr>
       <td><label>Host:</label></td>
       <td><a rel="nofollow" style='text-decoration:none;'>{{ .Host }}</a></td>
     </tr>
     <tr>
       <td><label>DNS/IP:</label></td>
       <td><a rel="nofollow" style='text-decoration:none;'>{{ .Target }}</a></td>
     </tr>
     <tr>
       <td><label>Check type:</label></td>
       <td>{{ .ServiceType }}</td>
     </tr>
     <tr>
       <td><label>Port:</label></td>
       <td>{{ .Port }}</td>
     </tr>  
   </table>   
 </body>
</html>
`

type TemplateData struct {
	Host        string
	Target      string
	ServiceType string
	Port        string
	FailMessage string
}

func (e *Email) emailBody() string {
	var body bytes.Buffer

	var tmplText string
	if e.failed {
		tmplText = emailTemplateCritical_ENG
	} else {
		tmplText = emailTemplateOK_ENG
	}

	tmplData := TemplateData{
		Host:        e.serviceInfo.Host,
		Target:      e.serviceInfo.Target,
		ServiceType: e.serviceInfo.ServiceTypeString(),
		Port:        tryParsePort(e.serviceInfo.Metadata),
		FailMessage: e.failedMsg,
	}

	tmpl, err := template.New("email").Parse(tmplText)
	if err != nil {
		fmt.Errorf("failed to parse template: %s", err.Error())
	} else {
		err = tmpl.Execute(&body, tmplData)
		if err != nil {
			fmt.Errorf("failed to execute template: %s", err.Error())
		}
	}

	return body.String()
}

func (e *Email) emailSubject() string {
	//get string for bool value
	status := stringStatus(e.failed)
	// example '[AlerTea] CRITICAL: myhost - tcp:84.12.34.54'
	subject := fmt.Sprintf("AlerTea: %s - %s ", status, e.serviceInfo.Host)

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

// try parse port number from service metadata
// in case of service without port, it will return empty string
func tryParsePort(metadata string) string {
	port := ""
	regExp := regexp.MustCompile(`"port": (\d*),`)

	result := regExp.FindStringSubmatch(metadata)
	if len(result) > 1 {
		port = fmt.Sprintf("%s", result[1])
	}

	return port
}
