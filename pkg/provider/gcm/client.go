package gcm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/dialogs/dialog-push-service/pkg/provider"
	"github.com/pkg/errors"
)

const ErrorCodeFailedToReadResponse = "FailedToReadResponse"

// Client (legacy/gcm)
// https://firebase.google.com/docs/cloud-messaging/http-server-ref
// Legacy FCM/GCM API (https://firebase.google.com/docs/cloud-messaging/migrate-v1):
// 1. copy server key from: https://console.firebase.google.com/project/_/settings/cloudmessaging/
// 2. add to request header: Authorization:key=<server key>
type Client struct {
	client *http.Client

	// count send retries
	retries int

	// authorization key:
	// https://firebase.google.com/docs/cloud-messaging/migrate-v1#before_2
	headerAuthorization string

	sandbox bool
}

func New(key []byte, isSandbox bool, retries int, timeout time.Duration) (*Client, error) {

	if timeout <= 0 {
		timeout = time.Second * 10
	}

	return &Client{
		headerAuthorization: "key=" + string(key),
		retries:             retries,
		sandbox:             isSandbox,
		client: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *Client) Sandbox() bool {
	return c.sandbox
}

func (c *Client) Send(ctx context.Context, message *Request) (retval *Response, err error) {

	req, err := c.newRequest(ctx)
	if err != nil {
		return nil, err
	}

	if c.sandbox && message != nil {
		message.DryRun = true
	}

	fnSend := func() (int, error) {
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

	pipe := provider.NewPipe(func(w io.Writer) error {
		// message format:
		// https://firebase.google.com/docs/cloud-messaging/http-server-ref#downstream-http-messages-json
		return json.NewEncoder(w).Encode(message)
	})
	defer pipe.Close()

	req.Body = pipe

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	retval := &Response{
		StatusCode: res.StatusCode,
	}

	// https://firebase.google.com/docs/cloud-messaging/http-server-ref#error-codes
	if res.StatusCode == 200 || res.StatusCode == 400 {
		if err := provider.DecodeJSONResponse(res.Body, retval); err != nil {
			outInfo, errEncode := provider.JSONWithoutSecrets(message)
			if errEncode != nil {
				outInfo = []byte(errEncode.Error())
			}
			return nil, errors.Wrap(err, "invalid gcm response: source: "+string(outInfo))
		}
	}

	return retval, nil
}

func (c *Client) newRequest(ctx context.Context) (*http.Request, error) {

	req, err := http.NewRequest(http.MethodPost, "https://fcm.googleapis.com/fcm/send", nil)
	if err != nil {
		return nil, err
	}

	// token format:
	// https://firebase.google.com/docs/cloud-messaging/migrate-v1#before_2
	req.Header.Set("Authorization", c.headerAuthorization)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", "-1")
	req = req.WithContext(ctx)

	return req, nil
}
