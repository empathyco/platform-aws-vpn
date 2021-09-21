package ovpn

const kClientConfigTemplate = `# OpenVPN client settings
# Configuration file for {{ .Certificate.Subject.String }}
# Expires on: {{ .Certificate.NotAfter.Format "02 Jan 06 15:04 MST" }}

dev tun
client
remote {{ env "PKI_DOMAIN" }}
remote-random-hostname
push-peer-info
explicit-exit-notify
route 172.16.0.0 255.240.0.0

remote-cert-tls server
tls-version-min 1.3 or-highest
verify-x509-name '{{ env "PKI_DOMAIN" }}' name
cipher AES-256-GCM
auth SHA256
verb 3

# Serial Number: {{ .Certificate.SerialNumber.Text 16 }}
<cert>
{{ printf "%s" (pemCert .Certificate) -}}
</cert>

{{ if .CrossCert -}}
<extra-certs>
{{ printf "%s" (pemCert .CrossCert) -}}
</extra-certs>
{{- end }}

# Key ID: {{ printf "%x" .Certificate.SubjectKeyId }}
<key>
%PRIVATEKEY%</key>

<ca>
{{ printf "%s" (pemCert .CACert) -}}
{{- if .PrevCACert -}}{{ printf "%s" (pemCert .PrevCACert) -}}{{- end -}}
</ca>

<tls-crypt>
{{ printf "%s" .StaticKey -}}
</tls-crypt>
`
