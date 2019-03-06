package serverapi

import (
	"bytes"
	"crypto/x509/pkix"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/empathybroker/aws-vpn/pkg/api"
	awsservices "github.com/empathybroker/aws-vpn/pkg/aws"
	"github.com/empathybroker/aws-vpn/pkg/ovpn"
	"github.com/empathybroker/aws-vpn/pkg/pki"
	log "github.com/sirupsen/logrus"
	jose "gopkg.in/square/go-jose.v2"
)

type configRequest struct {
	PublicKey jose.JSONWebKey `json:"publicKey"`
}

func apiServerConfig(w http.ResponseWriter, r *http.Request) {
	var request configRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err, "Invalid input")
		return
	}

	if !request.PublicKey.Valid() {
		api.ErrorResponse(w, http.StatusBadRequest, nil, "Invalid public key")
		return
	}

	dnsName := os.Getenv("PKI_DOMAIN")
	if dnsName == "" {
		api.ErrorResponse(w, http.StatusInternalServerError, nil, "Missing domain name")
		return
	}

	cert, err := apiPKI.CreateCertificate(r.Context(), request.PublicKey.Key,
		pkix.Name{CommonName: dnsName}, pki.ServerCert,
		pki.WithDuration(30*24*time.Hour), pki.WithDNS(dnsName))
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err, "Error signing certificate")
		return
	}

	configData := ovpn.ConfigData{
		Certificate: cert.Certificate,

		CACert:     apiPKI.GetCACert(r.Context()),
		PrevCACert: apiPKI.GetPrevCACert(r.Context()),
		CrossCert:  apiPKI.GetCrossCert(r.Context()),

		StaticKey: apiPKI.GetStaticKey(r.Context()),
	}

	var config bytes.Buffer
	if err := ovpn.GetServerConfig(&config, configData); err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err, "Error obtaining config")
		return
	}

	event := api.J{
		"event":   "server_cert",
		"success": true,
		"request": request,
		"cert":    cert,
	}

	if err := awsservices.PublishEvent(apiSNS, r.Context(), event); err != nil {
		log.WithError(err).Error("Error publishing event")
	}

	api.JsonResponse(w, http.StatusOK, api.J{
		"message": "OK",
		"config":  config.Bytes(),
	})
}
