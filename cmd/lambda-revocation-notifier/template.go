package main

import (
	"text/template"

	"github.com/empathybroker/aws-vpn/pkg/pki"
)

const bodyTemplate = `Hi,

The following VPN certificates are approaching their expiration time:
{{ range .Certificates }}
* {{ .Serial }} expires on {{ (.NotAfter.Format "02/01/2006 15:04 MST") }}
{{- end }}

Make sure you request new certificates before then, or you will be unable to connect to the VPN.

You can manage your certificates at {{ .AdminURL }}

If you have any questions or need support, please visit {{ .HelpURL }}

If you are receiving this email in error, please contact us.

Regards,
- {{ .Signature }}`

var tplEmail = template.Must(template.New("email_body").Parse(bodyTemplate))

type bodyData struct {
	Certificates []*pki.CertificateInfo

	AdminURL  string
	HelpURL   string
	Signature string
}
