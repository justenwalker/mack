package libmacaroon

import (
	"encoding/base64"
	"github.com/google/go-cmp/cmp"
	"io"
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
		t.Errorf("want (-), got (+)\n%v", cmp.Diff(m.String(), m2.String()))
	}
}

func TestV2JEncoding_allocs(t *testing.T) {
	var m mack.Macaroon
	bs, err := base64.URLEncoding.DecodeString("eyJ2IjoyLCJsIjoiaHR0cDovL2V4YW1wbGUub3JnLyIsImkiOiJrZXlpZCIsImMiOlt7ImkiOiJhY2NvdW50ID0gMzczNTkyODU1OSJ9LHsiaSI6InVzZXIgPSBhbGljZSJ9XSwiczY0IjoiUy1sbnpSNmd4ckpycjJwS2xPNmJCYkZZaHRvTHFGNk1RcWs4alE0U1h2dyJ9")
	if err != nil {
		t.Fatalf("base64.URLEncoding.DecodeString: %v", err)
	}
	var v2j V2J
	if err = v2j.DecodeMacaroon(bs, &m); err != nil {
		t.Fatalf("Encoding.DecodeMacaroon: %v", err)
	}
	var enc []byte
	allocs := testing.AllocsPerRun(100_000, func() {
		enc, err = v2j.EncodeMacaroon(&m)
		if err != nil {
			t.Fatalf("Encoding.EncodeMacaroon: %v", err)
		}
		_, _ = io.Discard.Write(enc)
	})
	const maxAllocs = 9
	if allocs > maxAllocs {
		writeHeapProfile(t)
		t.Fatalf("allocs = %d > %d", int(allocs), maxAllocs)
	}
}

func TestV2JDecoding_allocs(t *testing.T) {
	bs, err := base64.URLEncoding.DecodeString("eyJ2IjoyLCJsIjoiaHR0cDovL2V4YW1wbGUub3JnLyIsImkiOiJrZXlpZCIsImMiOlt7ImkiOiJhY2NvdW50ID0gMzczNTkyODU1OSJ9LHsiaSI6InVzZXIgPSBhbGljZSJ9XSwiczY0IjoiUy1sbnpSNmd4ckpycjJwS2xPNmJCYkZZaHRvTHFGNk1RcWs4alE0U1h2dyJ9")
	if err != nil {
		t.Fatalf("base64.URLEncoding.DecodeString: %v", err)
	}
	var v2j V2J
	var m mack.Macaroon
	allocs := testing.AllocsPerRun(100_000, func() {
		if err = v2j.DecodeMacaroon(bs, &m); err != nil {
			t.Fatalf("Encoding.DecodeMacaroon: %v", err)
		}
	})
	const maxAllocs = 28
	if allocs > maxAllocs {
		writeHeapProfile(t)
		t.Fatalf("allocs = %d > %d", int(allocs), maxAllocs)
	}
}
