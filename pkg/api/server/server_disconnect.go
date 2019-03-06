package serverapi

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/empathybroker/aws-vpn/pkg/api"
	awsservices "github.com/empathybroker/aws-vpn/pkg/aws"
	log "github.com/sirupsen/logrus"
)

type disconnectRequest struct {
	CommonName string `json:"common_name"`
	TrustedIP  net.IP `json:"trusted_ip"`

	Duration      int `json:"duration"`
	BytesSent     int `json:"bytes_sent"`
	BytesReceived int `json:"bytes_received"`

	ClientHWAddr   string `json:"client_hwaddr"`
	ClientPlatform string `json:"client_platform"`
	ClientVersion  string `json:"client_version"`
	ClientGUI      string `json:"client_gui"`
	ClientSSL      string `json:"client_ssl"`
}

func apiServerDisconnect(w http.ResponseWriter, r *http.Request) {
	var request disconnectRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err, "Invalid input")
		return
	}

	event := api.J{
		"event":   "client_disconnect",
		"request": request,
	}

	if err := awsservices.PublishEvent(apiSNS, r.Context(), event); err != nil {
		log.WithError(err).Error("Error publishing event")
	}

	api.JsonResponse(w, http.StatusOK, api.J{"message": "OK"})
}
