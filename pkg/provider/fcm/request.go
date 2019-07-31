package fcm

const (
	AndroidMessagePriorityNormal AndroidMessagePriority = "NORMAL"
	AndroidMessagePriorityHigh   AndroidMessagePriority = "HIGH"
)

// AndroidMessagePriority values:
// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#androidmessagepriority
type AndroidMessagePriority string

// Request format:
// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages/send#request-body
type Request struct {
	ValidateOnly bool    `json:"validate_only,omitempty"`
	Message      Message `json:"message"`
}

// Message format:
// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#Message
type Message struct {
	Name         string            `json:"name,omitempty"`
	Token        string            `json:"token,omitempty"`
	Topic        string            `json:"topic,omitempty"`
	Condition    string            `json:"condition,omitempty"`
	Data         map[string]string `json:"data,omitempty"`
	Notification *Notification     `json:"notification,omitempty"`
	Android      *AndroidConfig    `json:"android,omitempty"`
	Webpush      *WebpushConfig    `json:"webpush,omitempty"`
	Apns         *ApnsConfig       `json:"apns,omitempty"`
	FcmOptions   *FcmOptions       `json:"fcm_options,omitempty"`
}

// Notification format:
// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#notification
type Notification struct {
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
	Image string `json:"image,omitempty"`
}

// AndroidConfig format
// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#androidconfig
type AndroidConfig struct {
	CollapseKey           string                 `json:"collapse_key,omitempty"`
	Priority              AndroidMessagePriority `json:"priority,omitempty"`
	TTL                   string                 `json:"ttl,omitempty"`
	RestrictedPackageName string                 `json:"restricted_package_name,omitempty"`
	Data                  map[string]string      `json:"data,omitempty"`
	Notification          *AndroidNotification   `json:"Notification,omitempty"`
	FcmOptions            AndroidFcmOptions      `json:"fcm_options,omitempty"`
}

// AndroidNotification format
// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#androidnotification
type AndroidNotification struct {
	Title        string   `json:"title,omitempty"`
	Body         string   `json:"body,omitempty"`
	Icon         string   `json:"icon,omitempty"`
	Color        string   `json:"color,omitempty"`
	Sound        string   `json:"sound,omitempty"`
	Tag          string   `json:"tag,omitempty"`
	ClickAction  string   `json:"click_action,omitempty"`
	BodyLocKey   string   `json:"body_loc_key,omitempty"`
	BodyLocArgs  []string `json:"body_loc_args,omitempty"`
	TitleLocKey  string   `json:"title_loc_key,omitempty"`
	TitleLocArgs []string `json:"title_loc_args,omitempty"`
	ChannelID    string   `json:"channel_id,omitempty"`
	Image        string   `json:"image,omitempty"`
}

// AndroidFcmOptions format:
// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#androidfcmoptions
type AndroidFcmOptions struct {
	AnalyticsLabel string `json:"analytics_label"`
}

// WebpushConfig format:
// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#webpushconfig
type WebpushConfig struct {
	Headers      map[string]string      `json:"headers,omitempty"`
	Data         map[string]string      `json:"data,omitempty"`
	Notification map[string]interface{} `json:"notification,omitempty"`
	FcmOptions   *WebpushFcmOptions     `json:"fcm_options,omitempty"`
}

// WebpushFcmOptions format:
// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#webpushfcmoptions
type WebpushFcmOptions struct {
	Link string `json:"link,omitempty"`
}

// ApnsConfig format:
// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#apnsconfig
type ApnsConfig struct {
	Headers    map[string]string      `json:"headers,omitempty"`
	Payload    map[string]interface{} `json:"payload,omitempty"`
	FcmOptions *ApnsFcmOptions        `json:"fcm_options,omitempty"`
}

// ApnsFcmOptions foramt:
// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#apnsfcmoptions
type ApnsFcmOptions struct {
	AnalyticsLabel string `json:"analytics_label,omitempty"`
	Image          string `json:"image,omitempty"`
}

// FcmOptions format:
// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages#fcmoptions
type FcmOptions struct {
	AnalyticsLabel string `json:"analytics_label,omitempty"`
}
