package ovpn

const kServerConfigTemplate = `# OpenVPN server settings
# Expires on: {{ .Certificate.NotAfter.Format "02 Jan 06 15:04 MST" }}

dev tun
topology subnet
server 10.8.0.0 255.255.255.0
keepalive 10 60
client-to-client
push-peer-info
opt-verify
mlock

auth-gen-token
explicit-exit-notify
script-security 2
status-version 3
mute-replay-warnings
verb 4

# scripts
setenv AWS_REGION {{ env "AWS_REGION" }}
setenv PKI_API_ENDPOINT https://{{ env "PKI_DOMAIN" }}/api
tls-verify /usr/local/bin/ovpn-helper
client-connect /usr/local/bin/ovpn-helper
client-disconnect /usr/local/bin/ovpn-helper

# Authentication
tls-server
auth SHA256
cipher AES-256-GCM
remote-cert-tls client
tls-cert-profile preferred
tls-version-min 1.3 or-highest
x509-username-field ext:subjectAltName

# Serial Number {{ .Certificate.SerialNumber.Text 16 }}
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

dh none
`
