package libmacaroon

import (
	"encoding/base64"
	"io"
	"testing"

	"github.com/justenwalker/mack"
)

func TestV2Encoding(t *testing.T) {
	var m mack.Macaroon
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
	var m2 mack.Macaroon
	if err = (V2{}).DecodeMacaroon(enc, &m2); err != nil {
		t.Fatalf("Encoding.DecodeMacaroon: %v", err)
	}
	if !m.Equal(&m2) {
		t.Errorf("got %#v, want %#v", m, m2)
	}
}

func TestV2Encoding_allocs(t *testing.T) {
	var m mack.Macaroon
	bs, err := base64.URLEncoding.DecodeString("AgETaHR0cDovL2V4YW1wbGUub3JnLwIFa2V5aWQAAhRhY2NvdW50ID0gMzczNTkyODU1OQACDHVzZXIgPSBhbGljZQAABiBL6WfNHqDGsmuvakqU7psFsViG2guoXoxCqTyNDhJe_A==")
	if err != nil {
		t.Fatalf("base64.URLEncoding.DecodeString: %v", err)
	}
	var v2 V2
	var enc []byte
	if err = v2.DecodeMacaroon(bs, &m); err != nil {
		t.Fatalf("Encoding.DecodeMacaroon: %v", err)
	}
	allocs := testing.AllocsPerRun(100_000, func() {
		enc, err = v2.EncodeMacaroon(&m)
		if err != nil {
			t.Fatalf("Encoding.EncodeMacaroon: %v", err)
		}
		_, _ = io.Discard.Write(enc)
	})
	const maxAllocs = 3
	if allocs > maxAllocs {
		writeHeapProfile(t)
		t.Fatalf("allocs = %d > %d", int(allocs), maxAllocs)
	}
	if allocs < maxAllocs {
		t.Logf("allocs = %d < %d; consider lowering the maxAllocs", int(allocs), maxAllocs)
	}
}

func TestV2Decoding_allocs(t *testing.T) {
	bs, err := base64.URLEncoding.DecodeString("AgETaHR0cDovL2V4YW1wbGUub3JnLwIFa2V5aWQAAhRhY2NvdW50ID0gMzczNTkyODU1OQACDHVzZXIgPSBhbGljZQAABiBL6WfNHqDGsmuvakqU7psFsViG2guoXoxCqTyNDhJe_A==")
	if err != nil {
		t.Fatalf("base64.URLEncoding.DecodeString: %v", err)
	}
	var v2 V2
	var m mack.Macaroon
	allocs := testing.AllocsPerRun(100_000, func() {
		if err = v2.DecodeMacaroon(bs, &m); err != nil {
			t.Fatalf("Encoding.DecodeMacaroon: %v", err)
		}
	})
	const maxAllocs = 5
	if allocs > maxAllocs {
		writeHeapProfile(t)
		t.Fatalf("allocs = %d > %d", int(allocs), maxAllocs)
	}
	if allocs < maxAllocs {
		t.Logf("allocs = %d < %d; consider lowering the maxAllocs", int(allocs), maxAllocs)
	}
}
