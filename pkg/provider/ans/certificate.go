package ans

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

// oid info:
// https://images.apple.com/certificateauthority/pdf/Apple_WWDR_CPS_v1.20.pdf
// https://github.com/SilentCircle/apns_tools/blob/master/FakeAppleWWDRCA.cfg
var (
	OidPushDevelop    = asn1.ObjectIdentifier([]int{1, 2, 840, 113635, 100, 6, 3, 1})
	OidPushProduction = asn1.ObjectIdentifier([]int{1, 2, 840, 113635, 100, 6, 3, 2})
	OidVoIPTopics     = asn1.ObjectIdentifier([]int{1, 2, 840, 113635, 100, 6, 3, 6})
	OidVoIP           = asn1.ObjectIdentifier([]int{1, 2, 840, 113635, 100, 6, 3, 5})
)

func ExistOID(cert *tls.Certificate, oid asn1.ObjectIdentifier) (bool, error) {

	values, err := GetOIDValue(cert, oid)
	if err != nil {
		return false, err
	}

	return len(values) > 0, nil
}

func GetOIDValue(cert *tls.Certificate, oid asn1.ObjectIdentifier) ([][]byte, error) {

	retval := make([][]byte, 0)

	for _, c := range cert.Certificate {
		cList, err := x509.ParseCertificates(c)
		if err != nil {
			return nil, err
		}

		for _, c := range cList {
			for _, e := range c.Extensions {
				if e.Id.Equal(oid) {
					retval = append(retval, e.Value)
				}
			}
		}
	}

	return retval, nil
}

// GetTopics returns topics from certificate extension value
// Binary data format:
// <block start=0x30> <block size=0xD> <value start=0xc> <value size=0x2> <value byte 1> <value byte 2>
// <block start=0x30> <block size=0x3> <value start=0xc> <value size=0x1> <value byte 1>
// <value start=0xc> <value size=0x2> <value byte 1> <value byte 2>
func GetTopics(src []byte) ([]string, error) {

	type State int
	const (
		BlockStart uint8 = 0x30
		ValueStart uint8 = 0xc

		NextBlock State = iota
		ReadBlockSize
		ReadValueSize
		ReadValue
	)

	var (
		r         = bytes.NewReader(src)
		value     = bytes.NewBuffer(nil)
		state     = NextBlock
		retval    = make([]string, 0)
		valueSize byte
	)

	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		switch state {
		case NextBlock:
			if b == BlockStart {
				state = ReadBlockSize
			} else if b == ValueStart {
				state = ReadValueSize
			} else {
				return nil, fmt.Errorf("topic: unknown block ID: %v", b)
			}

		case ReadBlockSize:
			state = NextBlock

		case ReadValueSize:
			valueSize = b
			state = ReadValue

		case ReadValue:
			value.WriteByte(b)
			valueSize--

			if valueSize == 0 {
				retval = append(retval, value.String())
				value.Reset()
				state = NextBlock
			}
		}
	}

	return retval, nil
}

func CheckVoIPTopicByCert(topic string, cert *tls.Certificate) error {

	oidValues, err := GetOIDValue(cert, OidVoIPTopics)
	if err != nil {
		return errors.Wrap(err, "read VoIP topics")
	}

	topicList := make([]string, 0, 10)
	for _, value := range oidValues {
		list, err := GetTopics(value)
		if err != nil {
			return err
		}

		topicList = append(topicList, list...)
	}

	// check config topic by certificate
	if topic != "" {
		var exist bool
		for i := range topicList {
			if strings.Compare(topicList[i], topic) == 0 {
				exist = true
				break
			}
		}

		if !exist {
			return fmt.Errorf("invalid VoIP topic: '%s' (topics in certificate: %v)", topic, topicList)
		}
	}

	return nil
}
