package fcm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

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
	sendEndpoint string

	// count send tries
	sendTries int

	// oauth token
	token atomic.Value

	// service-account config:
	// https://console.firebase.google.com/project/_/settings/serviceaccounts/adminsdk
	jwtConfig *jwt.Config
}

func New(serviceAccount []byte, sendTries int, timeout time.Duration) (*Client, error) {

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

	return &Client{
		sendEndpoint: getSendEndpoint(account.ProjectID),
		sendTries:    sendTries,
		jwtConfig:    jwtConfig,
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

	for try := 0; try < sendTries; try++ {
		retval, err = c.send(ctx, message)
		if err != nil {
			return nil, err
		}

		if retval.Ok() || !retval.Error.IsRemoteError() {
			break
		}
	}

	return retval, nil
}

func (c *Client) send(ctx context.Context, message *Request) (*Response, error) {

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
		return nil, err
	}

	res, err := c.client.Do(req)
	if err == nil {
		err = <-sendBodyErr
		defer res.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	retval := &Response{}
	if err := json.NewDecoder(res.Body).Decode(retval); err != nil {
		return nil, err
	}

	return retval, nil
}

func (c *Client) newRequest(ctx context.Context, body io.ReadCloser) (*http.Request, error) {

	token, err := c.getToken(ctx)
	if err != nil {
		return nil, err
	}

	// message format:
	// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages/send#request-body
	req, err := http.NewRequest(http.MethodPost, c.sendEndpoint, nil)
	if err != nil {
		return nil, err
	}

	// token format:
	// https://firebase.google.com/docs/cloud-messaging/auth-server
	req.Header.Set("Authorization", token.Type()+" "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", "-1")
	req = req.WithContext(ctx)

	req.Body = body
	req.GetBody = func() (io.ReadCloser, error) {
		return body, nil
	}

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

func getSendEndpoint(projectID string) string {

	projectID = url.PathEscape(projectID)
	return "https://fcm.googleapis.com/v1/projects/" + projectID + "/messages:send"
}
