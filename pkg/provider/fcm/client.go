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

	// count send retries
	retries int

	// oauth token
	token atomic.Value

	// service-account config:
	// https://console.firebase.google.com/project/_/settings/serviceaccounts/adminsdk
	jwtConfig *jwt.Config

	sandbox bool
}

func New(serviceAccount []byte, isSandbox bool, retries int, timeout time.Duration) (*Client, error) {

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

	if timeout <= 0 {
		timeout = time.Second * 10
	}

	return &Client{
		endpoint:  getEndpoint(account.ProjectID),
		retries:   retries,
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

	payload, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	fnSend := func() (statusCode int, _ error) {
		var e error
		retval, e = c.send(ctx, req, payload)
		if e != nil {
			return 0, e
		}

		return retval.StatusCode, nil
	}

	err = provider.SendWithRetry(c.retries, fnSend)
	if err != nil {
		return nil, err
	}

	return retval, nil
}

func (c *Client) send(ctx context.Context, req *http.Request, message json.RawMessage) (*Response, error) {

	{
		pipe := provider.NewPipe(func(w io.Writer) (err error) {
			// Request format:
			// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages/send#request-body
			return json.NewEncoder(w).Encode(&Request{
				ValidateOnly: c.sandbox,
				Message:      message,
			})
		})
		defer pipe.Close()

		token, err := c.getToken(ctx)
		if err != nil {
			return nil, err
		}

		// token format:
		// https://firebase.google.com/docs/cloud-messaging/auth-server
		req.Header.Set("Authorization", token.Type()+" "+token.AccessToken)

		req.Body = pipe
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	retval := &Response{
		StatusCode: res.StatusCode,
	}

	// https://firebase.google.com/docs/reference/fcm/rest/v1/ErrorCode
	switch retval.StatusCode {
	case 200, 400, 401, 403, 404, 429:
		if err := provider.DecodeJSONResponse(res.Body, retval); err != nil {
			outInfo, errEncode := provider.JSONWithoutSecrets(message)
			if errEncode != nil {
				outInfo = []byte(errEncode.Error())
			}
			return nil, errors.Wrap(err, "invalid fcm response: source: "+string(outInfo))
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
