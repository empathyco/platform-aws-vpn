package clientapi

import (
	"net/http"

	"github.com/empathybroker/aws-vpn/pkg/api"
	awsservices "github.com/empathybroker/aws-vpn/pkg/aws"
	"github.com/empathybroker/aws-vpn/pkg/pki"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func apiRevokeCert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	serial, err := pki.DecodeSerial(vars["serial"])
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err, "Invalid serial")
		return
	}

	_, userInfo, err := api.GetAPIGWPrincipal(r)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err, "Error obtaining principal")
		return
	}

	cert, err := apiPKI.GetCertBySerial(r.Context(), serial)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err, "Error obtaining certificate")
		return
	}

	if cert == nil || (!userInfo.IsAdmin && cert.Subject != userInfo.Email) {
		api.ErrorResponse(w, http.StatusNotFound, nil, "Not Found")
		return
	}

	cert, err = apiPKI.RevokeCert(r.Context(), serial)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err, "Error revoking certificate")
		return
	}

	if cert == nil {
		api.ErrorResponse(w, http.StatusNotFound, nil, "Not found")
		return
	}

	event := api.J{
		"event":      "cert_revoked",
		"revoked_by": userInfo.Email,
		"cert":       cert,
	}

	if err := awsservices.PublishEvent(apiSNS, r.Context(), event); err != nil {
		log.WithError(err).Error("Error publishing event")
	}

	api.JsonResponse(w, http.StatusOK, cert)
}
