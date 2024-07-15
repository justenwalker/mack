package libmacaroon

import (
	"encoding/base64"
	"testing"

	"github.com/justenwalker/mack/macaroon"
)

func TestV1JEncoding(t *testing.T) {
	var m macaroon.Macaroon
	bs, err := base64.URLEncoding.DecodeString("eyJsb2NhdGlvbiI6Imh0dHA6Ly9leGFtcGxlLm9yZy8iLCJpZGVudGlmaWVyIjoia2V5aWQiLCJjYXZlYXRzIjpbeyJjaWQiOiJhY2NvdW50ID0gMzczNTkyODU1OSJ9LHsiY2lkIjoidXNlciA9IGFsaWNlIn1dLCJzaWduYXR1cmUiOiI0YmU5NjdjZDFlYTBjNmIyNmJhZjZhNGE5NGVlOWIwNWIxNTg4NmRhMGJhODVlOGM0MmE5M2M4ZDBlMTI1ZWZjIn0=")
	if err != nil {
		t.Fatalf("base64.URLEncoding.DecodeString: %v", err)
	}
	if err = (V1J{}).DecodeMacaroon(bs, &m); err != nil {
		t.Fatalf("Encoding.DecodeMacaroon: %v", err)
	}
	var enc []byte
	if enc, err = (V1J{}).EncodeMacaroon(&m); err != nil {
		t.Fatalf("Encoding.EncodeMacaroon: %v", err)
	}
	var m2 macaroon.Macaroon
	if err = (V1J{}).DecodeMacaroon(enc, &m2); err != nil {
		t.Fatalf("Encoding.DecodeMacaroon: %v", err)
	}
	if !m.Equal(&m2) {
		t.Errorf("got %#v, want %#v", m, m2)
	}
}
