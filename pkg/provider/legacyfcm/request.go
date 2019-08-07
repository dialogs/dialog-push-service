package legacyfcm

import "encoding/json"

type Request struct {
	To                    string                 `json:"to"`
	RegistrationIDs       []string               `json:"registration_ids,omitempty"`
	Condition             string                 `json:"condition,omitempty"`
	NotificationKey       string                 `json:"notification_key,omitempty"`
	CollapseKey           string                 `json:"collapse_key,omitempty"`
	Priority              string                 `json:"priority,omitempty"`
	ContentAvailable      bool                   `json:"content_available,omitempty"`
	MutableContent        json.RawMessage        `json:"mutable_content,omitempty"`
	TimeToLive            int                    `json:"time_to_live,omitempty"`
	RestrictedPackageName string                 `json:"restricted_package_name,omitempty"`
	DryRun                bool                   `json:"dry_run,omitempty"`
	Data                  map[string]interface{} `json:"data,omitempty"`
	Notification          *Notification          `json:"notification,omitempty"`
}

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
