package gcm

import "encoding/json"

// https://firebase.google.com/docs/cloud-messaging/http-server-ref#notification-payload-support
type Notification struct {
	Title            string          `json:"title,omitempty"`
	Body             string          `json:"body,omitempty"`
	AndroidChannelID string          `json:"android_channel_id,omitempty"`
	Icon             string          `json:"icon,omitempty"`
	Sound            string          `json:"sound,omitempty"`
	Tag              string          `json:"tag,omitempty"`
	Color            string          `json:"color,omitempty"`
	ClickAction      string          `json:"click_action,omitempty"`
	BodyLocKey       string          `json:"body_loc_key,omitempty"`
	BodyLocArgs      json.RawMessage `json:"body_loc_args,omitempty"`
	TitleLocKey      string          `json:"title_loc_key,omitempty"`
	TitleLocArgs     json.RawMessage `json:"title_loc_args,omitempty"`
}

// https://firebase.google.com/docs/cloud-messaging/http-server-ref#downstream-http-messages-json
type Request struct {
	To                    string          `json:"to"`
	RegistrationIDs       []string        `json:"registration_ids,omitempty"`
	Condition             string          `json:"condition,omitempty"`
	NotificationKey       string          `json:"notification_key,omitempty"`
	CollapseKey           string          `json:"collapse_key,omitempty"`
	Priority              string          `json:"priority,omitempty"`
	ContentAvailable      bool            `json:"content_available,omitempty"`
	TimeToLive            int             `json:"time_to_live,omitempty"`
	RestrictedPackageName string          `json:"restricted_package_name,omitempty"`
	DryRun                bool            `json:"dry_run,omitempty"`
	Data                  json.RawMessage `json:"data,omitempty"`
	MutableContent        json.RawMessage `json:"mutable_content,omitempty"`
	Notification          json.RawMessage `json:"notification,omitempty"`
}

func (r *Request) SetToken(token string) {
	if r != nil {
		r.To = token
	}
}

func (r *Request) ShouldIgnore() bool {
	return r == nil
}
