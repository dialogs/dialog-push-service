package ans

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/asn1"

	"github.com/pkg/errors"
	"github.com/sideshow/apns2"
)

var oidPushDevelop = asn1.ObjectIdentifier([]int{1, 2, 840, 113635, 100, 6, 3, 1})

type Client struct {
	native *apns2.Client
}

func New(certTLS *tls.Certificate) (*Client, error) {

	isDevelopCert, err := isDevelopCert(certTLS)
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

func (c *Client) Send(ctx context.Context, req *Request) (*Response, error) {

	nativeReq := req.native()

	nativeRes, err := c.native.PushWithContext(ctx, nativeReq)
	if err != nil {
		return nil, errors.Wrap(err, "push request")
	}

	res := NewResponse(nativeRes)

	return res, nil
}

func isDevelopCert(cert *tls.Certificate) (bool, error) {

	for _, c := range cert.Certificate {
		cList, err := x509.ParseCertificates(c)
		if err != nil {
			return false, err
		}

		for _, c := range cList {
			for _, e := range c.Extensions {
				if e.Id.Equal(oidPushDevelop) {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
