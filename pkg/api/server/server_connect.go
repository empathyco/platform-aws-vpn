package serverapi

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/empathybroker/aws-vpn/pkg/api"
	awsservices "github.com/empathybroker/aws-vpn/pkg/aws"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type connectRequest struct {
	CommonName string `json:"common_name"`
	TrustedIP  net.IP `json:"trusted_ip"`

	ClientHWAddr   string `json:"client_hwaddr"`
	ClientPlatform string `json:"client_platform"`
	ClientVersion  string `json:"client_version"`
	ClientGUI      string `json:"client_gui"`
	ClientSSL      string `json:"client_ssl"`
}

func subnetToRoute(subnetStr string) (string, error) {
	if _, subnet, err := net.ParseCIDR(subnetStr); err == nil {
		mask := net.IPv4(subnet.Mask[0], subnet.Mask[1], subnet.Mask[2], subnet.Mask[3])
		return fmt.Sprintf("route %s %s", subnet.IP.String(), mask.String()), nil
	} else {
		return "", errors.Wrapf(err, "parsing subnet CIDR: %s", subnetStr)
	}
}

func apiServerConnect(w http.ResponseWriter, r *http.Request) {
	var request connectRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err, "Invalid input")
		return
	}

	userInfo, err := apiDirectory.GetUserInfo(r.Context(), request.CommonName)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err, "Could not get user info")
		return
	}

	if userInfo.IsSuspended {
		event := api.J{
			"event":   "client_connect",
			"success": false,
			"error":   "user_suspended",
			"request": request,
		}

		if err := awsservices.PublishEvent(apiSNS, r.Context(), event); err != nil {
			log.WithError(err).Error("Error publishing event")
		}

		api.ErrorResponse(w, http.StatusUnauthorized, nil, "User is suspended")
		return
	}

	var push []string
	if vpnSchema, ok := userInfo.Schemas["VPN"]; ok {
		if macaddrs, ok := vpnSchema["Allowed_MACs"].([]interface{}); ok && len(macaddrs) > 0 {

			found := false
			for _, macaddrI := range macaddrs {
				if macaddr, ok := macaddrI.(string); ok {
					if macaddr == request.ClientHWAddr {
						found = true
						break
					}
				}
			}

			if !found {
				event := api.J{
					"event":    "client_connect",
					"success":  false,
					"error":    "mac_mismatch",
					"expected": macaddrs,
					"request":  request,
				}

				if err := awsservices.PublishEvent(apiSNS, r.Context(), event); err != nil {
					log.WithError(err).Error("Error publishing event")
				}

				api.ErrorResponse(w, http.StatusUnauthorized, nil, "Unexpected MAC address")
				return
			}
		}

		if subnets, ok := vpnSchema["Allowed_subnets"].([]interface{}); ok {
			for _, subnetI := range subnets {
				if subnetStr, ok := subnetI.(string); ok {
					if route, err := subnetToRoute(subnetStr); err == nil {
						push = append(push, route)
					} else {
						log.WithError(err).Error("Error parsing route from user schema")
					}
				}
			}
		}
	} else {
		// error missing schema
	}

	if os.Getenv("PKI_ROUTE_EC2_PREFIX_LIST") == "true" {
		if pl, err := apiEC2.DescribePrefixListsWithContext(r.Context(), &ec2.DescribePrefixListsInput{}); err == nil {
			for _, p := range pl.PrefixLists {
				for _, subnet := range aws.StringValueSlice(p.Cidrs) {
					if route, err := subnetToRoute(subnet); err == nil {
						push = append(push, route)
					} else {
						log.WithError(err).Error("Error parsing route from prefix list")
					}
				}
			}
		}
	}

	if domainSearch := os.Getenv("PKI_DOMAIN_SEARCH"); domainSearch != "" {
		for _, domain := range strings.Split(domainSearch, ",") {
			push = append(push, fmt.Sprintf("dhcp-option DOMAIN %s", domain))
		}
	}

	push = append(push, "dhcp-option DNS 10.8.0.1")
	push = append(push, "route-metric 101")

	event := api.J{
		"event":   "client_connect",
		"success": true,
		"request": request,
		"push":    push,
	}

	if err := awsservices.PublishEvent(apiSNS, r.Context(), event); err != nil {
		log.WithError(err).Error("Error publishing event")
	}

	api.JsonResponse(w, http.StatusOK, api.J{
		"message": "OK",
		"push":    push,
	})
}
