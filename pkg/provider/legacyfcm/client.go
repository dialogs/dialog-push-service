package legacyfcm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"
)

const ErrorCodeFailedToReadResponse = "FailedToReadResponse"

// Client (legacy)
// https://firebase.google.com/docs/cloud-messaging/http-server-ref
// Legacy FCM/GCM API (https://firebase.google.com/docs/cloud-messaging/migrate-v1):
// 1. copy server key from: https://console.firebase.google.com/project/_/settings/cloudmessaging/
// 2. add to request header: Authorization:key=<server key>
type Client struct {
	client *http.Client

	// count send tries
	sendTries int

	// authorization key:
	// https://firebase.google.com/docs/cloud-messaging/migrate-v1#before_2
	headerAuthorization string
}

func New(key string, sendTries int, timeout time.Duration) (*Client, error) {

	return &Client{
		headerAuthorization: "key=" + key,
		sendTries:           sendTries,
		client: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *Client) Send(ctx context.Context, message *Request) (retval *Response, err error) {

	sendTries := c.sendTries
	if sendTries <= 0 {
		sendTries = 1
	}

	var statusCode int

	for try := 0; try < sendTries; try++ {
		statusCode, retval, err = c.send(ctx, message)
		if err != nil {
			existTries := try < sendTries-1
			if existTries &&
				(statusCode == http.StatusInternalServerError || err == context.DeadlineExceeded) {
				continue
			}

			return nil, err
		}

		if statusCode == http.StatusOK && len(retval.Results) > 0 {

			var retry bool
			for _, item := range retval.Results {
				if item.Error == ErrorCodeUnavailable {
					retry = true
					break
				}
			}

			if retry {
				continue
			}
		}

		break
	}

	return retval, nil
}

func (c *Client) send(ctx context.Context, message *Request) (int, *Response, error) {

	sendBodyErr := make(chan error, 1)

	reqBodyReader, reqBodyWriter := io.Pipe()
	defer reqBodyReader.Close()

	go func() {
		defer reqBodyWriter.Close()
		defer close(sendBodyErr)

		sendBodyErr <- json.NewEncoder(reqBodyWriter).Encode(message)
	}()

	req, err := c.newRequest(ctx, reqBodyReader)
	if err != nil {
		return 0, nil, err
	}

	res, err := c.client.Do(req)
	if err == nil {
		err = <-sendBodyErr
		defer res.Body.Close()
	}

	if err != nil {
		return 0, nil, err
	}

	if res.StatusCode != 200 {
		return res.StatusCode, nil, errors.New(strconv.Itoa(res.StatusCode) + " " + res.Status)
	}

	retval := &Response{}
	if err := json.NewDecoder(res.Body).Decode(retval); err != nil {
		return res.StatusCode, nil, err
	}

	retval.StatusCode = res.StatusCode

	return res.StatusCode, retval, nil
}

func (c *Client) newRequest(ctx context.Context, body io.ReadCloser) (*http.Request, error) {

	// message format:
	// https://firebase.google.com/docs/cloud-messaging/http-server-ref#downstream-http-messages-json
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

	req.Body = body
	req.GetBody = func() (io.ReadCloser, error) {
		return body, nil
	}

	return req, nil
}
