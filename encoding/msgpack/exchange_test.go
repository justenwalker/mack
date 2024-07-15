package msgpack

import (
	"bytes"
	"testing"
	"testing/quick"

	"github.com/google/go-cmp/cmp"

	"github.com/justenwalker/mack/exchange"
	"github.com/justenwalker/mack/macaroon/thirdparty"
)

func TestEncodeMessage_quick(t *testing.T) {
	fn := func(kt string, kid string, payload []byte) bool {
		in := exchange.EncryptedMessage{
			Type:    kt,
			KeyID:   kid,
			Payload: payload,
		}
		bs, err := Encoding.EncodeMessage(&in)
		if err != nil {
			t.Errorf("EncodeMessage: %v", err)
			return false
		}
		out, err := Encoding.DecodeMessage(bs)
		if err != nil {
			t.Errorf("DecodeMessage: %v", err)
			return false
		}
		if diff := cmp.Diff(in, *out, cmp.Comparer(compareBytes)); diff != "" {
			t.Errorf("DecodeMessage returned diff (-want +got):\n%s", diff)
			return false
		}
		return true
	}
	if err := quick.Check(fn, nil); err != nil {
		t.Fatal(err)
	}
}

func TestEncodeTicket_quick(t *testing.T) {
	fn := func(cK []byte, predicate []byte) bool {
		in := thirdparty.Ticket{
			CaveatKey: cK,
			Predicate: predicate,
		}
		bs, err := Encoding.EncodeTicket(in)
		if err != nil {
			t.Errorf("EncodeTicket: %v", err)
			return false
		}
		out, err := Encoding.DecodeTicket(bs)
		if err != nil {
			t.Errorf("DecodeTicket: %v", err)
			return false
		}
		if diff := cmp.Diff(in, *out, cmp.Comparer(compareBytes)); diff != "" {
			t.Errorf("DecodeTicket returned diff (-want +got):\n%s", diff)
			return false
		}
		return true
	}
	if err := quick.Check(fn, nil); err != nil {
		t.Fatal(err)
	}
}

func compareBytes(a []byte, b []byte) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	return bytes.Equal(a, b)
}
