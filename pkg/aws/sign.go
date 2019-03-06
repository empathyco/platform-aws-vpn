package awsservices

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
)

type awsSigningTransport struct {
	parent        http.RoundTripper
	signer        *v4.Signer
	serviceName   string
	serviceRegion string
}

func NewAWSSigner(s *session.Session, serviceName string, parent http.RoundTripper) *awsSigningTransport {
	return &awsSigningTransport{
		signer:        v4.NewSigner(s.Config.Credentials),
		serviceRegion: aws.StringValue(s.Config.Region),
		serviceName:   serviceName,
		parent:        parent,
	}
}

func (t *awsSigningTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = cloneRequest(req)

	var body io.ReadSeeker
	if req.Body != nil {
		defer req.Body.Close()
		payload, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		body = bytes.NewReader(payload)
		req.Body = ioutil.NopCloser(body)
	}

	_, err := t.signer.Sign(req, body, t.serviceName, t.serviceRegion, time.Now())
	if err != nil {
		return nil, err
	}

	return t.parent.RoundTrip(req)
}

// CloneRequest creates a shallow copy of the request along with a deep copy of the Headers.
func cloneRequest(req *http.Request) *http.Request {
	r := new(http.Request)

	// shallow clone
	*r = *req

	// deep copy headers
	r.Header = cloneHeader(req.Header)

	return r
}

// CloneHeader creates a deep copy of an http.Header.
func cloneHeader(in http.Header) http.Header {
	out := make(http.Header, len(in))
	for key, values := range in {
		newValues := make([]string, len(values))
		copy(newValues, values)
		out[key] = newValues
	}
	return out
}
