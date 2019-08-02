package ans

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/asn1"
	"net/url"

	"github.com/pkg/errors"
	"github.com/sideshow/apns2"
)

// oid info:
// https://images.apple.com/certificateauthority/pdf/Apple_WWDR_CPS_v1.20.pdf
// https://github.com/SilentCircle/apns_tools/blob/master/FakeAppleWWDRCA.cfg
var (
	OidPushDevelop = asn1.ObjectIdentifier([]int{1, 2, 840, 113635, 100, 6, 3, 1})
	OidVoIPTopics  = asn1.ObjectIdentifier([]int{1, 2, 840, 113635, 100, 6, 3, 4})
	OidVoIP        = asn1.ObjectIdentifier([]int{1, 2, 840, 113635, 100, 6, 3, 5})
)

type Client struct {
	native *apns2.Client
}

func New(certTLS *tls.Certificate) (*Client, error) {

	isDevelopCert, err := ExistOID(certTLS, OidPushDevelop)
	if err != nil {
		return nil, errors.Wrap(err, "check certificate type")
	}

	native := apns2.NewClient(*certTLS)
	if isDevelopCert {
		native.Development()
	} else {
		native.Production()
	}

	return &Client{
		native: native,
	}, nil
}

func NewFromPem(pem []byte) (*Client, error) {

	certTLS, err := tls.X509KeyPair(pem, pem)
	if err != nil {
		return nil, errors.Wrap(err, "read certificate")
	}

	return New(&certTLS)
}

func (c *Client) Certificate() tls.Certificate {
	return c.native.Certificate
}

func (c *Client) Send(ctx context.Context, req *Request) (*Response, error) {

	nativeReq := req.native()

	nativeRes, err := c.native.PushWithContext(ctx, nativeReq)
	if err != nil {
		if urlError, ok := err.(*url.Error); ok {
			// hide device token in the error info
			// original error:
			// Post https://api.development.push.apple.com/3/device/<token>: dial tcp: lookup api.development.push.apple.com: no such host
			return nil, urlError.Err
		}

		return nil, err
	}

	res := NewResponse(nativeRes)

	return res, nil
}

func ExistOID(cert *tls.Certificate, oid asn1.ObjectIdentifier) (bool, error) {

	values, err := GetOIDValue(cert, oid)
	if err != nil {
		return false, err
	}

	return len(values) > 0, nil
}

func GetOIDValue(cert *tls.Certificate, oid asn1.ObjectIdentifier) ([][]byte, error) {

	retval := make([][]byte, 0)

	for _, c := range cert.Certificate {
		cList, err := x509.ParseCertificates(c)
		if err != nil {
			return nil, err
		}

		for _, c := range cList {
			for _, e := range c.Extensions {
				if e.Id.Equal(oid) {
					retval = append(retval, e.Value)
				}
			}
		}
	}

	return retval, nil
}
