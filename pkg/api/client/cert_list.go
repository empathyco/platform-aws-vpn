package clientapi

import (
	"net/http"

	"github.com/empathybroker/aws-vpn/pkg/api"
)

func apiGetCerts(w http.ResponseWriter, r *http.Request) {
	_, userInfo, err := api.GetAPIGWPrincipal(r)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err, "Error obtaining principal")
		return
	}

	subject := userInfo.Email
	if userInfo.IsAdmin && r.URL.Query().Get("all") == "true" {
		subject = ""
	}

	certs, err := apiPKI.ListCerts(r.Context(), subject)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err, "Error listing certs")
		return
	}

	api.JsonResponse(w, http.StatusOK, api.J{
		"isAdmin": userInfo.IsAdmin,
		"certs":   certs,
	})
}
