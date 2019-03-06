package clientapi

import (
	"net/http"

	"github.com/empathybroker/aws-vpn/pkg/api"

	"github.com/empathybroker/aws-vpn/pkg/pki"
	"github.com/gorilla/mux"
)

func apiGetCert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	serial, err := pki.DecodeSerial(vars["serial"])
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err, "Invalid serial")
		return
	}

	cert, err := apiPKI.GetCertBySerial(r.Context(), serial)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err, "Error obtaining certificate")
		return
	}

	_, userInfo, err := api.GetAPIGWPrincipal(r)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err, "Error obtaining principal")
		return
	}

	if cert == nil || cert.Subject != userInfo.Email {
		api.ErrorResponse(w, http.StatusNotFound, nil, "Not found")
		return
	}

	api.JsonResponse(w, http.StatusOK, cert)
}
