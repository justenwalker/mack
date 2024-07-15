package msgpack

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/justenwalker/mack/internal/testhelpers"
	"github.com/justenwalker/mack/macaroon"
)

func TestEncodeDecodeMacaroon(t *testing.T) {
	fx := testhelpers.CreateTestFixture(t, testhelpers.FixtureConfig{
		ID:       "test",
		Location: "https://example.com/fp",
		Caveats: []testhelpers.Caveat{
			{ID: "a > 1"},
			{
				ID:         "user = foo",
				ThirdParty: "https://example.com/3p",
			},
			{ID: "b < 2"},
		},
	})
	encoded, err := Encoding.EncodeMacaroon(fx.Target)
	if err != nil {
		t.Fatalf("failed to encode stack: %s", err)
	}
	var decoded macaroon.Macaroon
	err = Encoding.DecodeMacaroon(encoded, &decoded)
	if err != nil {
		t.Fatalf("failed to decode stack: %s", err)
	}
	if diff := cmp.Diff(fx.Target, &decoded,
		cmp.Comparer(macaroonCompare),
		cmp.Comparer(macaroonComparePointers)); diff != "" {
		t.Errorf("decoded macaroon mismatch (-want +got):\n%s", diff)
	}
}

func TestEncodeDecodeStack(t *testing.T) {
	fx := testhelpers.CreateTestFixture(t, testhelpers.FixtureConfig{
		ID:       "test",
		Location: "https://example.com/fp",
		Caveats: []testhelpers.Caveat{
			{ID: "a > 1"},
			{
				ID:         "user = foo",
				ThirdParty: "3p",
			},
			{ID: "b < 2"},
		},
	})
	encoded, err := Encoding.EncodeStack(fx.Stack)
	if err != nil {
		t.Fatalf("failed to encode stack: %s", err)
	}
	var decoded macaroon.Stack
	err = Encoding.DecodeStack(encoded, &decoded)
	if err != nil {
		t.Fatalf("failed to decode stack: %s", err)
	}
	if diff := cmp.Diff(fx.Stack, decoded,
		cmp.Comparer(macaroonCompare),
		cmp.Comparer(macaroonComparePointers)); diff != "" {
		t.Errorf("decoded stack mismatch (-want +got):\n%s", diff)
	}
}

func macaroonCompare(a macaroon.Macaroon, b macaroon.Macaroon) bool {
	return a.Equal(&b)
}

func macaroonComparePointers(a *macaroon.Macaroon, b *macaroon.Macaroon) bool {
	return a.Equal(b)
}
