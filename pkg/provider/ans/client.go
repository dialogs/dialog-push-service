package ans

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/dialogs/dialog-push-service/pkg/provider"
	"github.com/pkg/errors"
	"golang.org/x/net/http2"
)

type Client struct {
	client         *http.Client
	endpointPrefix string
	certTLS        tls.Certificate
	sandbox        bool
	retries        int
	supportsVoIP   bool
}

func New(certTLS *tls.Certificate, isSandbox bool, retries int, timeout time.Duration) (*Client, error) {

	hasDevelopCert, err := ExistOID(certTLS, OidPushDevelop)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read certificate property 'Apple Push Notification service Development'")
	}

	hasProductionCert, err := ExistOID(certTLS, OidPushProduction)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read certificate property 'Apple Push Notification service Production'")
	}

	supportsVoIP, err := ExistOID(certTLS, OidVoIP)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read certificate property 'supports VoIP'")
	}

	sandbox := hasDevelopCert && (!hasProductionCert || isSandbox)

	endpointPrefix := "https://api.push.apple.com"
	if sandbox {
		endpointPrefix = "https://api.development.push.apple.com"
	}
	endpointPrefix += "/3/device/"

	if timeout <= 0 {
		timeout = time.Second * 10
	}

	client := newHttpClient(certTLS, timeout)

	return &Client{
		client:         client,
		endpointPrefix: endpointPrefix,
		certTLS:        *certTLS,
		sandbox:        sandbox,
		retries:        retries,
		supportsVoIP:   supportsVoIP,
	}, nil
}

func NewFromPem(pemData []byte, isSandbox bool, retries int, timeout time.Duration) (*Client, error) {

	certTLS, err := tls.X509KeyPair(pemData, pemData)
	if err != nil {
		return nil, errors.Wrap(err, "read certificate")
	}

	return New(&certTLS, isSandbox, retries, timeout)
}

func (c *Client) Certificate() tls.Certificate {
	return c.certTLS
}

func (c *Client) Sandbox() bool {
	return c.sandbox
}

func (c *Client) SupportsVoIP() bool {
	return c.supportsVoIP
}

func (c *Client) Send(ctx context.Context, message *Request) (retval *Response, err error) {

	req, err := c.newRequest(ctx, message.Token, &message.Headers)
	if err != nil {
		return nil, err
	}

	fnSend := func() (statusCode int, _ error) {
		retval, err = c.send(ctx, req, message)
		if err != nil {
			return 0, err
		}

		return retval.StatusCode, err
	}

	err = provider.SendWithRetry(c.retries, fnSend)
	if err != nil {
		return nil, err
	}

	return retval, nil
}

func (c *Client) send(ctx context.Context, req *http.Request, message *Request) (*Response, error) {

	body := ioutil.NopCloser(bytes.NewReader(message.Payload))
	req.Body = body

	res, err := c.client.Do(req)
	if err != nil {
		if urlError, ok := err.(*url.Error); ok {
			// hide device token in the error info
			// original error:
			// Post https://api.development.push.apple.com/3/device/<token>: dial tcp: lookup api.development.push.apple.com: no such host
			return nil, urlError.Err
		}

		return nil, err
	}
	defer res.Body.Close()

	resp := NewResponse(res.Header.Get("apns-id"), res.StatusCode)
	switch resp.StatusCode {
	case 200, 400, 403, 404, 405, 410, 413, 429, 500, 503:
		// Table 8-6Values for the APNs JSON reason key
		// https://developer.apple.com/library/archive/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/CommunicatingwithAPNs.html#//apple_ref/doc/uid/TP40008194-CH11-SW1
		if err := provider.DecodeJSONResponse(res.Body, &resp.Body); err != nil && err != io.EOF {
			out, errEncode := provider.JSONWithoutSecrets(message)
			if errEncode != nil {
				out = []byte(errEncode.Error())
			}
			return nil, errors.Wrap(err, "invalid asn response: source: "+string(out))
		}
	}

	return resp, nil
}

func (c *Client) newRequest(ctx context.Context, token string, header *RequestHeader) (*http.Request, error) {

	endpoint := c.endpointPrefix + token
	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	// Table 8-2 APNs request headers
	// https://developer.apple.com/library/archive/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/CommunicatingwithAPNs.html#//apple_ref/doc/uid/TP40008194-CH11-SW1
	if header.ID != "" {
		req.Header.Set("apns-id", header.ID)
	}

	if !header.Expiration.IsZero() {
		req.Header.Set("apns-expiration", strconv.FormatInt(header.Expiration.Unix(), 10))
	}

	if header.Priority > 0 {
		req.Header.Set("apns-priority", strconv.Itoa(header.Priority))
	}

	if header.Topic != "" {
		req.Header.Set("apns-topic", header.Topic)
	}

	if header.CollapseID != "" {
		req.Header.Set("apns-collapse-id", header.CollapseID)
	}

	req = req.WithContext(ctx)

	return req, nil
}

func newHttpClient(certTLS *tls.Certificate, timeout time.Duration) *http.Client {

	dial := func(network, addr string, cfg *tls.Config) (net.Conn, error) {
		dialer := &net.Dialer{
			Timeout:   timeout,
			KeepAlive: timeout,
		}
		return tls.DialWithDialer(dialer, network, addr, cfg)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*certTLS},
	}
	if len(certTLS.Certificate) > 0 {
		tlsConfig.BuildNameToCertificate()
	}

	transport := &http2.Transport{
		TLSClientConfig: tlsConfig,
		DialTLS:         dial,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}
