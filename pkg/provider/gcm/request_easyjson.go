// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package gcm

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson3c9d2b01DecodeGithubComDialogsDialogPushServicePkgProviderGcm(in *jlexer.Lexer, out *Request) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "to":
			out.To = string(in.String())
		case "registration_ids":
			if in.IsNull() {
				in.Skip()
				out.RegistrationIDs = nil
			} else {
				in.Delim('[')
				if out.RegistrationIDs == nil {
					if !in.IsDelim(']') {
						out.RegistrationIDs = make([]string, 0, 4)
					} else {
						out.RegistrationIDs = []string{}
					}
				} else {
					out.RegistrationIDs = (out.RegistrationIDs)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.RegistrationIDs = append(out.RegistrationIDs, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "condition":
			out.Condition = string(in.String())
		case "notification_key":
			out.NotificationKey = string(in.String())
		case "collapse_key":
			out.CollapseKey = string(in.String())
		case "priority":
			out.Priority = string(in.String())
		case "content_available":
			out.ContentAvailable = bool(in.Bool())
		case "mutable_content":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.MutableContent).UnmarshalJSON(data))
			}
		case "time_to_live":
			out.TimeToLive = int(in.Int())
		case "restricted_package_name":
			out.RestrictedPackageName = string(in.String())
		case "dry_run":
			out.DryRun = bool(in.Bool())
		case "data":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.Data).UnmarshalJSON(data))
			}
		case "notification":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.Notification).UnmarshalJSON(data))
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson3c9d2b01EncodeGithubComDialogsDialogPushServicePkgProviderGcm(out *jwriter.Writer, in Request) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"to\":"
		out.RawString(prefix[1:])
		out.String(string(in.To))
	}
	if len(in.RegistrationIDs) != 0 {
		const prefix string = ",\"registration_ids\":"
		out.RawString(prefix)
		{
			out.RawByte('[')
			for v2, v3 := range in.RegistrationIDs {
				if v2 > 0 {
					out.RawByte(',')
				}
				out.String(string(v3))
			}
			out.RawByte(']')
		}
	}
	if in.Condition != "" {
		const prefix string = ",\"condition\":"
		out.RawString(prefix)
		out.String(string(in.Condition))
	}
	if in.NotificationKey != "" {
		const prefix string = ",\"notification_key\":"
		out.RawString(prefix)
		out.String(string(in.NotificationKey))
	}
	if in.CollapseKey != "" {
		const prefix string = ",\"collapse_key\":"
		out.RawString(prefix)
		out.String(string(in.CollapseKey))
	}
	if in.Priority != "" {
		const prefix string = ",\"priority\":"
		out.RawString(prefix)
		out.String(string(in.Priority))
	}
	if in.ContentAvailable {
		const prefix string = ",\"content_available\":"
		out.RawString(prefix)
		out.Bool(bool(in.ContentAvailable))
	}
	if len(in.MutableContent) != 0 {
		const prefix string = ",\"mutable_content\":"
		out.RawString(prefix)
		out.Raw((in.MutableContent).MarshalJSON())
	}
	if in.TimeToLive != 0 {
		const prefix string = ",\"time_to_live\":"
		out.RawString(prefix)
		out.Int(int(in.TimeToLive))
	}
	if in.RestrictedPackageName != "" {
		const prefix string = ",\"restricted_package_name\":"
		out.RawString(prefix)
		out.String(string(in.RestrictedPackageName))
	}
	if in.DryRun {
		const prefix string = ",\"dry_run\":"
		out.RawString(prefix)
		out.Bool(bool(in.DryRun))
	}
	if len(in.Data) != 0 {
		const prefix string = ",\"data\":"
		out.RawString(prefix)
		out.Raw((in.Data).MarshalJSON())
	}
	if len(in.Notification) != 0 {
		const prefix string = ",\"notification\":"
		out.RawString(prefix)
		out.Raw((in.Notification).MarshalJSON())
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Request) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson3c9d2b01EncodeGithubComDialogsDialogPushServicePkgProviderGcm(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Request) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson3c9d2b01EncodeGithubComDialogsDialogPushServicePkgProviderGcm(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Request) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson3c9d2b01DecodeGithubComDialogsDialogPushServicePkgProviderGcm(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Request) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson3c9d2b01DecodeGithubComDialogsDialogPushServicePkgProviderGcm(l, v)
}
func easyjson3c9d2b01DecodeGithubComDialogsDialogPushServicePkgProviderGcm1(in *jlexer.Lexer, out *Notification) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "title":
			out.Title = string(in.String())
		case "body":
			out.Body = string(in.String())
		case "android_channel_id":
			out.AndroidChannelID = string(in.String())
		case "icon":
			out.Icon = string(in.String())
		case "sound":
			out.Sound = string(in.String())
		case "tag":
			out.Tag = string(in.String())
		case "color":
			out.Color = string(in.String())
		case "click_action":
			out.ClickAction = string(in.String())
		case "body_loc_key":
			out.BodyLocKey = string(in.String())
		case "body_loc_args":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.BodyLocArgs).UnmarshalJSON(data))
			}
		case "title_loc_key":
			out.TitleLocKey = string(in.String())
		case "title_loc_args":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.TitleLocArgs).UnmarshalJSON(data))
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson3c9d2b01EncodeGithubComDialogsDialogPushServicePkgProviderGcm1(out *jwriter.Writer, in Notification) {
	out.RawByte('{')
	first := true
	_ = first
	if in.Title != "" {
		const prefix string = ",\"title\":"
		first = false
		out.RawString(prefix[1:])
		out.String(string(in.Title))
	}
	if in.Body != "" {
		const prefix string = ",\"body\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Body))
	}
	if in.AndroidChannelID != "" {
		const prefix string = ",\"android_channel_id\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.AndroidChannelID))
	}
	if in.Icon != "" {
		const prefix string = ",\"icon\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Icon))
	}
	if in.Sound != "" {
		const prefix string = ",\"sound\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Sound))
	}
	if in.Tag != "" {
		const prefix string = ",\"tag\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Tag))
	}
	if in.Color != "" {
		const prefix string = ",\"color\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Color))
	}
	if in.ClickAction != "" {
		const prefix string = ",\"click_action\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.ClickAction))
	}
	if in.BodyLocKey != "" {
		const prefix string = ",\"body_loc_key\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.BodyLocKey))
	}
	if len(in.BodyLocArgs) != 0 {
		const prefix string = ",\"body_loc_args\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Raw((in.BodyLocArgs).MarshalJSON())
	}
	if in.TitleLocKey != "" {
		const prefix string = ",\"title_loc_key\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.TitleLocKey))
	}
	if len(in.TitleLocArgs) != 0 {
		const prefix string = ",\"title_loc_args\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Raw((in.TitleLocArgs).MarshalJSON())
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Notification) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson3c9d2b01EncodeGithubComDialogsDialogPushServicePkgProviderGcm1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Notification) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson3c9d2b01EncodeGithubComDialogsDialogPushServicePkgProviderGcm1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Notification) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson3c9d2b01DecodeGithubComDialogsDialogPushServicePkgProviderGcm1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Notification) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson3c9d2b01DecodeGithubComDialogsDialogPushServicePkgProviderGcm1(l, v)
}
