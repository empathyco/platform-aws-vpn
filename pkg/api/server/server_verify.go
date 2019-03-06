package serverapi

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/empathybroker/aws-vpn/pkg/api"
	awsservices "github.com/empathybroker/aws-vpn/pkg/aws"
	"github.com/empathybroker/aws-vpn/pkg/pki"
	log "github.com/sirupsen/logrus"
)

type verifyRequest struct {
	Subject string `json:"subject"`
	Serial  string `json:"serial"`
	Digest  string `json:"digest"`
	Client  net.IP `json:"client"`
}

func apiServerVerify(w http.ResponseWriter, r *http.Request) {
	var request verifyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err, "Invalid input")
		return
	}

	serial, err := pki.DecodeSerial(request.Serial)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err, "Invalid serial")
		return
	}

	cert, err := apiPKI.GetCertBySerial(r.Context(), serial)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err, "Error fetching certificate")
		return
	}

	if cert == nil {
		event := api.J{
			"event":   "cert_verify",
			"success": false,
			"error":   "cert_not_found",
			"request": request,
		}

		if err := awsservices.PublishEvent(apiSNS, r.Context(), event); err != nil {
			log.WithError(err).Error("Error publishing event")
		}

		api.ErrorResponse(w, http.StatusUnauthorized, err, "Certificate not found")
		return
	}

	if cert.Revoked != nil {
		event := api.J{
			"event":   "cert_verify",
			"success": false,
			"error":   "cert_revoked",
			"request": request,
		}

		if err := awsservices.PublishEvent(apiSNS, r.Context(), event); err != nil {
			log.WithError(err).Error("Error publishing event")
		}

		api.ErrorResponse(w, http.StatusForbidden, err, "Certificate has been revoked")
		return
	}

	event := api.J{
		"event":   "cert_verify",
		"success": true,
		"request": request,
		"cert":    cert,
	}

	if err := awsservices.PublishEvent(apiSNS, r.Context(), event); err != nil {
		log.WithError(err).Error("Error publishing event")
	}

	api.JsonResponse(w, http.StatusOK, api.J{"message": "OK"})
}
