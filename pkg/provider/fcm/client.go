package fcm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/dialogs/dialog-push-service/pkg/provider"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

// Google API clients:
// https://github.com/firebase/firebase-admin-go
// https://github.com/googleapis/google-api-go-client

// New FCM API:
// 1. download service-account.json from https://console.firebase.google.com/project/_/settings/serviceaccounts/adminsdk
// 2. add to request header: Authorization: Bearer <oauth token>

// Legacy FCM/GCM API (https://firebase.google.com/docs/cloud-messaging/migrate-v1):
// 1. copy server key from: https://console.firebase.google.com/project/_/settings/cloudmessaging/
// 2. add to request header: Authorization:key=<server key>

type Client struct {
	client *http.Client

	// send message endpoint:
	// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages/send
	endpoint string

	// count send tries
	sendTries int

	// oauth token
	token atomic.Value

	// service-account config:
	// https://console.firebase.google.com/project/_/settings/serviceaccounts/adminsdk
	jwtConfig *jwt.Config

	sandbox bool
}

func New(serviceAccount []byte, isSandbox bool, sendTries int, timeout time.Duration) (*Client, error) {

	scope := []string{
		// To authorize access to FCM, request:
		// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages/send#authorization-scopes
		"https://www.googleapis.com/auth/firebase.messaging",
	}

	jwtConfig, err := google.JWTConfigFromJSON(serviceAccount, scope...)
	if err != nil {
		return nil, errors.Wrap(err, "jwt config")
	}

	account := &struct {
		ProjectID string `json:"project_id"`
	}{}

	if err := json.Unmarshal(serviceAccount, account); err != nil {
		return nil, errors.Wrap(err, "account")
	}

	if sendTries <= 0 {
		sendTries = 2
	}

	if timeout <= 0 {
		timeout = time.Second * 10
	}

	return &Client{
		endpoint:  getEndpoint(account.ProjectID),
		sendTries: sendTries,
		jwtConfig: jwtConfig,
		sandbox:   isSandbox,
		client: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *Client) Sandbox() bool {
	return c.sandbox
}

func (c *Client) Send(ctx context.Context, message *Message) (retval *Response, err error) {

	req, err := c.newRequest(ctx)
	if err != nil {
		return nil, err
	}

	payload, err := message.MarshalJSON()
	if err != nil {
		return nil, err
	}

	fnSend := func() (statusCode int, _ error) {
		retval, err = c.send(ctx, req, payload)
		if err != nil {
			return 0, err
		}

		return retval.StatusCode, err
	}

	err = provider.SendWithRetry(c.sendTries, fnSend)
	if err != nil {
		return nil, err
	}

	return retval, nil
}

var (
	_RequestBodyPrefixForSandbox = []byte(`{"validate_only":true,"message":`)
	_RequestBodyPrefix           = []byte(`{"message":`)
	_RequestBodySuffix           = []byte(`}`)
)

func (c *Client) send(ctx context.Context, req *http.Request, message json.RawMessage) (*Response, error) {

	sendBodyErr := make(chan error, 1)

	{
		reqBodyReader, reqBodyWriter := io.Pipe()
		defer reqBodyReader.Close()

		go func() {
			defer reqBodyWriter.Close()
			defer close(sendBodyErr)

			var err error

			// Request format:
			// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages/send#request-body
			if c.sandbox {
				_, err = reqBodyWriter.Write(_RequestBodyPrefixForSandbox)
			} else {
				_, err = reqBodyWriter.Write(_RequestBodyPrefix)
			}

			if err == nil {
				_, err = reqBodyWriter.Write(message)
			}

			if err == nil {
				_, err = reqBodyWriter.Write(_RequestBodySuffix)
			}

			if err != nil {
				sendBodyErr <- err
				return
			}
		}()

		token, err := c.getToken(ctx)
		if err != nil {
			return nil, err
		}

		// token format:
		// https://firebase.google.com/docs/cloud-messaging/auth-server
		req.Header.Set("Authorization", token.Type()+" "+token.AccessToken)

		req.Body = reqBodyReader
		req.GetBody = func() (io.ReadCloser, error) {
			return reqBodyReader, nil
		}
	}

	res, err := c.client.Do(req)
	if err == nil {
		err = <-sendBodyErr
		defer res.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	retval := &Response{
		StatusCode: res.StatusCode,
	}

	// https://firebase.google.com/docs/reference/fcm/rest/v1/ErrorCode
	switch retval.StatusCode {
	case 200, 400, 401, 403, 404, 429:
		if err := json.NewDecoder(res.Body).Decode(retval); err != nil {
			return nil, err
		}
	}

	return retval, nil
}

func (c *Client) newRequest(ctx context.Context) (*http.Request, error) {

	req, err := http.NewRequest(http.MethodPost, c.endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", "-1")
	req = req.WithContext(ctx)

	return req, nil
}

func (c *Client) getToken(ctx context.Context) (*oauth2.Token, error) {

	src := c.token.Load()
	if src != nil {
		token := src.(*oauth2.Token)
		if token.Valid() {
			return token, nil
		}
	}

	// source:
	// https://github.com/googleapis/google-api-go-client/blob/0c3fc9a1ae141ce9db158d15b06bca77ddcb923b/google-api-go-generator/gen.go#L613
	token, err := c.jwtConfig.TokenSource(ctx).Token()
	if err != nil {
		return nil, errors.Wrap(err, "jwt token")
	}

	c.token.Store(token)

	return token, nil
}

func getEndpoint(projectID string) string {

	projectID = url.PathEscape(projectID)
	return "https://fcm.googleapis.com/v1/projects/" + projectID + "/messages:send"
}
