package libmacaroon

import (
	"encoding/base64"
	"testing"

	"github.com/justenwalker/mack"
)

func TestV2JEncoding(t *testing.T) {
	var m mack.Macaroon
	bs, err := base64.URLEncoding.DecodeString("eyJ2IjoyLCJsIjoiaHR0cDovL2V4YW1wbGUub3JnLyIsImkiOiJrZXlpZCIsImMiOlt7ImkiOiJhY2NvdW50ID0gMzczNTkyODU1OSJ9LHsiaSI6InVzZXIgPSBhbGljZSJ9XSwiczY0IjoiUy1sbnpSNmd4ckpycjJwS2xPNmJCYkZZaHRvTHFGNk1RcWs4alE0U1h2dyJ9")
	if err != nil {
		t.Fatalf("base64.URLEncoding.DecodeString: %v", err)
	}
	if err = (V2J{}).DecodeMacaroon(bs, &m); err != nil {
		t.Fatalf("Encoding.DecodeMacaroon: %v", err)
	}
	var enc []byte
	if enc, err = (V2J{}).EncodeMacaroon(&m); err != nil {
		t.Fatalf("Encoding.EncodeMacaroon: %v", err)
	}
	var m2 mack.Macaroon
	if err = (V2J{}).DecodeMacaroon(enc, &m2); err != nil {
		t.Fatalf("Encoding.DecodeMacaroon: %v", err)
	}
	if !m.Equal(&m2) {
		t.Errorf("got %#v, want %#v", m, m2)
	}
}
