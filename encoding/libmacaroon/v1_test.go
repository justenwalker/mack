package libmacaroon

import (
	"encoding/base64"
	"testing"

	"github.com/justenwalker/mack/macaroon"
)

func TestV1Encoding(t *testing.T) {
	bs, err := base64.StdEncoding.DecodeString("TURBeU1XeHZZMkYwYVc5dUlHaDBkSEE2THk5bGVHRnRjR3hsTG05eVp5OEtNREF4Tldsa1pXNTBhV1pwWlhJZ2EyVjVhV1FLTURBeFpHTnBaQ0JoWTJOdmRXNTBJRDBnTXpjek5Ua3lPRFUxT1Fvd01ERTFZMmxrSUhWelpYSWdQU0JoYkdsalpRb3dNREptYzJsbmJtRjBkWEpsSUV2cFo4MGVvTWF5YTY5cVNwVHVtd1d4V0liYUM2aGVqRUtwUEkwT0VsNzhDZw==")
	if err != nil {
		t.Fatalf("base64.URLEncoding.DecodeString: %v", err)
	}
	b64 := &Base64{Encoding: base64.RawURLEncoding}
	v1 := V1{
		OutputEncoder: b64,
		InputDecoder:  b64,
	}
	var m macaroon.Macaroon
	if err = v1.DecodeMacaroon(bs, &m); err != nil {
		t.Fatalf("Encoding.DecodeMacaroon: %v", err)
	}
	var enc []byte
	if enc, err = v1.EncodeMacaroon(&m); err != nil {
		t.Fatalf("Encoding.EncodeMacaroon: %v", err)
	}
	var m2 macaroon.Macaroon
	if err = v1.DecodeMacaroon(enc, &m2); err != nil {
		t.Fatalf("Encoding.DecodeMacaroon: %v", err)
	}
	if !m.Equal(&m2) {
		t.Errorf("got %#v, want %#v", m, m2)
	}
}

func TestV1_DecodeStack(t *testing.T) {
	bs, err := base64.StdEncoding.DecodeString("TURBeU1XeHZZMkYwYVc5dUlHaDBkSEE2THk5bGVHRnRjR3hsTG05eVp5OEtNREF4Tldsa1pXNTBhV1pwWlhJZ2EyVjVhV1FLTURBeFpHTnBaQ0JoWTJOdmRXNTBJRDBnTXpjek5Ua3lPRFUxT1Fvd01ERTFZMmxrSUhWelpYSWdQU0JoYkdsalpRb3dNREptYzJsbmJtRjBkWEpsSUV2cFo4MGVvTWF5YTY5cVNwVHVtd1d4V0liYUM2aGVqRUtwUEkwT0VsNzhDZw==")
	if err != nil {
		t.Fatalf("base64.URLEncoding.DecodeString: %v", err)
	}
	b64 := &Base64{Encoding: base64.RawURLEncoding}
	v1 := V1{
		OutputEncoder: b64,
		InputDecoder:  b64,
	}
	var m macaroon.Macaroon
	if err = v1.DecodeMacaroon(bs, &m); err != nil {
		t.Fatalf("Encoding.DecodeMacaroon: %v", err)
	}

	var enc []byte
	stack := macaroon.Stack{m, m, m}
	if enc, err = v1.EncodeStack(stack); err != nil {
		t.Fatalf("Encoding.EncodeStack: %v", err)
	}
	var stack2 macaroon.Stack
	if err = v1.DecodeStack(enc, &stack2); err != nil {
		t.Fatalf("Encoding.DecodeStack: %v", err)
	}
	t.Log(stack2)
}