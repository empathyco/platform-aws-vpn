package clientapi

import (
	"net/http"

	"github.com/aws/aws-xray-sdk-go/xray"
	awsservices "github.com/empathybroker/aws-vpn/pkg/aws"
	"github.com/empathybroker/aws-vpn/pkg/pki"
	awspki "github.com/empathybroker/aws-vpn/pkg/pki/aws"
	"github.com/gorilla/mux"
)

var (
	pkiStorage = awspki.NewAWSStorage(awsservices.NewSecretsManagerClient(), awsservices.NewDynamoDBClient())
	apiPKI     = pki.NewPKI(pkiStorage)
	apiSNS     = awsservices.NewSNSClient()
)

func NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.Use(func(handler http.Handler) http.Handler {
		return xray.Handler(xray.NewFixedSegmentNamer("vpn-api-client"), handler)
	})

	r.HandleFunc("/certificates", apiGetCerts).Methods(http.MethodGet)
	r.HandleFunc("/certificates", apiNewCert).Methods(http.MethodPut)
	r.HandleFunc("/certificates/{serial}", apiGetCert).Methods(http.MethodGet)
	r.HandleFunc("/certificates/{serial}", apiRevokeCert).Methods(http.MethodDelete)

	return r
}
