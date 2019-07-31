package legacyfcm

import (
	"github.com/edganiukov/fcm"
)

// Client (legacy)
// https://firebase.google.com/docs/cloud-messaging/http-server-ref
type Client struct {
	native    *fcm.Client
	sendTries int
}

func New(key string, sendTries int) (*Client, error) {

	native, err := fcm.NewClient(key, fcm.WithEndpoint(fcm.DefaultEndpoint))
	if err != nil {
		return nil, err
	}

	return &Client{
		native:    native,
		sendTries: sendTries,
	}, nil
}

func (c *Client) Send(msq *fcm.Message) (*fcm.Response, error) {
	return c.native.SendWithRetry(msq, c.sendTries)
}
