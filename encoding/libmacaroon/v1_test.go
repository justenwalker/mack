package libmacaroon

import (
	"encoding/base64"
	"io"
	"testing"

	"github.com/justenwalker/mack"
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
	var m mack.Macaroon
	if err = v1.DecodeMacaroon(bs, &m); err != nil {
		t.Fatalf("Encoding.DecodeMacaroon: %v", err)
	}
	var enc []byte
	if enc, err = v1.EncodeMacaroon(&m); err != nil {
		t.Fatalf("Encoding.EncodeMacaroon: %v", err)
	}
	var m2 mack.Macaroon
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
	var m mack.Macaroon
	if err = v1.DecodeMacaroon(bs, &m); err != nil {
		t.Fatalf("Encoding.DecodeMacaroon: %v", err)
	}

	var enc []byte
	stack := mack.Stack{m, m, m}
	if enc, err = v1.EncodeStack(stack); err != nil {
		t.Fatalf("Encoding.EncodeStack: %v", err)
	}
	var stack2 mack.Stack
	if err = v1.DecodeStack(enc, &stack2); err != nil {
		t.Fatalf("Encoding.DecodeStack: %v", err)
	}
	t.Log(stack2)
}

func TestV1Encoding_allocs(t *testing.T) {
	bs, err := base64.StdEncoding.DecodeString("TURBeU1XeHZZMkYwYVc5dUlHaDBkSEE2THk5bGVHRnRjR3hsTG05eVp5OEtNREF4Tldsa1pXNTBhV1pwWlhJZ2EyVjVhV1FLTURBeFpHTnBaQ0JoWTJOdmRXNTBJRDBnTXpjek5Ua3lPRFUxT1Fvd01ERTFZMmxrSUhWelpYSWdQU0JoYkdsalpRb3dNREptYzJsbmJtRjBkWEpsSUV2cFo4MGVvTWF5YTY5cVNwVHVtd1d4V0liYUM2aGVqRUtwUEkwT0VsNzhDZw==")
	if err != nil {
		t.Fatalf("base64.URLEncoding.DecodeString: %v", err)
	}
	b64 := &Base64{Encoding: base64.RawURLEncoding}
	v1 := V1{
		InputDecoder: b64,
	}
	var m mack.Macaroon
	var enc []byte
	if err = v1.DecodeMacaroon(bs, &m); err != nil {
		t.Fatalf("Encoding.DecodeMacaroon: %v", err)
	}
	allocs := testing.AllocsPerRun(100_000, func() {
		enc, err = v1.EncodeMacaroon(&m)
		if err != nil {
			t.Fatalf("Encoding.EncodeMacaroon: %v", err)
		}
		_, _ = io.Discard.Write(enc)
	})
	const maxAllocs = 10
	if allocs > maxAllocs {
		writeHeapProfile(t)
		t.Fatalf("allocs = %d > %d", int(allocs), maxAllocs)
	}
}
