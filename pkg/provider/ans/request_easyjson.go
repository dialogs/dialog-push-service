// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package ans

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

func easyjson3c9d2b01DecodeGithubComDialogsDialogPushServicePkgProviderAns(in *jlexer.Lexer, out *RequestHeader) {
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
		case "id":
			out.ID = string(in.String())
		case "expiration":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.Expiration).UnmarshalJSON(data))
			}
		case "priority":
			out.Priority = int(in.Int())
		case "topic":
			out.Topic = string(in.String())
		case "collapse-id":
			out.CollapseID = string(in.String())
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
func easyjson3c9d2b01EncodeGithubComDialogsDialogPushServicePkgProviderAns(out *jwriter.Writer, in RequestHeader) {
	out.RawByte('{')
	first := true
	_ = first
	if in.ID != "" {
		const prefix string = ",\"id\":"
		first = false
		out.RawString(prefix[1:])
		out.String(string(in.ID))
	}
	if true {
		const prefix string = ",\"expiration\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Raw((in.Expiration).MarshalJSON())
	}
	if in.Priority != 0 {
		const prefix string = ",\"priority\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int(int(in.Priority))
	}
	if in.Topic != "" {
		const prefix string = ",\"topic\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Topic))
	}
	if in.CollapseID != "" {
		const prefix string = ",\"collapse-id\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.CollapseID))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v RequestHeader) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson3c9d2b01EncodeGithubComDialogsDialogPushServicePkgProviderAns(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v RequestHeader) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson3c9d2b01EncodeGithubComDialogsDialogPushServicePkgProviderAns(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *RequestHeader) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson3c9d2b01DecodeGithubComDialogsDialogPushServicePkgProviderAns(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *RequestHeader) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson3c9d2b01DecodeGithubComDialogsDialogPushServicePkgProviderAns(l, v)
}
func easyjson3c9d2b01DecodeGithubComDialogsDialogPushServicePkgProviderAns1(in *jlexer.Lexer, out *Request) {
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
		case "token":
			out.Token = string(in.String())
		case "headers":
			(out.Headers).UnmarshalEasyJSON(in)
		case "payload":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.Payload).UnmarshalJSON(data))
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
func easyjson3c9d2b01EncodeGithubComDialogsDialogPushServicePkgProviderAns1(out *jwriter.Writer, in Request) {
	out.RawByte('{')
	first := true
	_ = first
	if in.Token != "" {
		const prefix string = ",\"token\":"
		first = false
		out.RawString(prefix[1:])
		out.String(string(in.Token))
	}
	if true {
		const prefix string = ",\"headers\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		(in.Headers).MarshalEasyJSON(out)
	}
	if len(in.Payload) != 0 {
		const prefix string = ",\"payload\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Raw((in.Payload).MarshalJSON())
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Request) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson3c9d2b01EncodeGithubComDialogsDialogPushServicePkgProviderAns1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Request) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson3c9d2b01EncodeGithubComDialogsDialogPushServicePkgProviderAns1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Request) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson3c9d2b01DecodeGithubComDialogsDialogPushServicePkgProviderAns1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Request) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson3c9d2b01DecodeGithubComDialogsDialogPushServicePkgProviderAns1(l, v)
}
