package libmacaroon

import (
	"encoding/base64"
	"testing"

	"github.com/justenwalker/mack/macaroon"
)

func TestV2Encoding(t *testing.T) {
	var m macaroon.Macaroon
	bs, err := base64.URLEncoding.DecodeString("AgETaHR0cDovL2V4YW1wbGUub3JnLwIFa2V5aWQAAhRhY2NvdW50ID0gMzczNTkyODU1OQACDHVzZXIgPSBhbGljZQAABiBL6WfNHqDGsmuvakqU7psFsViG2guoXoxCqTyNDhJe_A==")
	if err != nil {
		t.Fatalf("base64.URLEncoding.DecodeString: %v", err)
	}
	if err = (V2{}).DecodeMacaroon(bs, &m); err != nil {
		t.Fatalf("Encoding.DecodeMacaroon: %v", err)
	}
	var enc []byte
	t.Logf("%v", &m)
	if enc, err = (V2{}).EncodeMacaroon(&m); err != nil {
		t.Fatalf("Encoding.EncodeMacaroon: %v", err)
	}
	var m2 macaroon.Macaroon
	if err = (V2{}).DecodeMacaroon(enc, &m2); err != nil {
		t.Fatalf("Encoding.DecodeMacaroon: %v", err)
	}
	if !m.Equal(&m2) {
		t.Errorf("got %#v, want %#v", m, m2)
	}
}
