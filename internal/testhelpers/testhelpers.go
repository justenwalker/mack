package testhelpers

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/justenwalker/mack/macaroon"
)

//go:generate go run go.uber.org/mock/mockgen -source ../../macaroon/verify.go -destination mocks.go -typed -package testhelpers

type Fixture struct {
	Scheme    *macaroon.Scheme
	Target    *macaroon.Macaroon
	Discharge []macaroon.Macaroon
	Stack     macaroon.Stack
}

type FixtureConfig struct {
	ID       string
	Location string
	Caveats  []Caveat
	Debug    bool
}

type Caveat struct {
	ID         string
	ThirdParty string
	Caveats    []Caveat
}

var (
	RootKey       = []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}
	ThirdPartyKey = []byte{2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, 1}
)

func CreateTestFixture(tb testing.TB, cfg FixtureConfig) Fixture {
	tb.Helper()
	var fx Fixture
	fx.Scheme = NewScheme(tb)
	var err error
	if cfg.Location == "" {
		cfg.Location = "1p"
	}
	fx.Target, fx.Discharge = makeMacaroon(tb, fx.Scheme, cfg.Location, cfg.ID, RootKey, cfg.Caveats, cfg.Debug)
	fx.Stack, err = fx.Scheme.PrepareStack(fx.Target, fx.Discharge)
	if err != nil {
		tb.Fatalf("ts.PrepareStack(%v,%v) = %v", fx.Target, fx.Discharge, err)
	}
	return fx
}

func makeMacaroon(tb testing.TB, s *macaroon.Scheme, loc string, id string, key []byte, cs []Caveat, debug bool) (*macaroon.Macaroon, []macaroon.Macaroon) {
	tb.Helper()
	m, err := s.UnsafeRootMacaroon(loc, []byte(id), key)
	if err != nil {
		tb.Fatalf("UnsafeRootMacaroon(%s,%s,%s) = %v", loc, id, hex.EncodeToString(key), err)
	}
	tp := make(map[string]struct{})
	var discharges []macaroon.Macaroon
	for _, c := range cs {
		if c.ThirdParty == "" {
			m, _ = s.AddFirstPartyCaveat(&m, []byte(c.ID))
			continue
		}
		if _, ok := tp[c.ID]; ok {
			continue
		}

		tp[c.ThirdParty] = struct{}{}
		thirdPartyDischarge, additionalDischarge := makeMacaroon(tb, s, c.ThirdParty, c.ID, ThirdPartyKey, c.Caveats, debug)
		additionalDischarge = append([]macaroon.Macaroon{*thirdPartyDischarge}, additionalDischarge...)
		discharges = append(discharges, additionalDischarge...)

		m, err = s.AddThirdPartyCaveat(&m, ThirdPartyKey, []byte(c.ID), c.ThirdParty)
		if err != nil {
			tb.Fatalf("AddThirdPartyCaveat(%s,%s,%s) = %v", hex.EncodeToString(ThirdPartyKey), c.ID, c.ThirdParty, err)
		}
	}
	if debug {
		tb.Logf("macaroon: %s", m.String())
		for i := range discharges {
			tb.Logf("ds[%d]: %s", i, discharges[i].String())
		}
	}
	return &m, discharges
}

func printableString(bs []byte) string {
	if len(bs) == 0 {
		return ""
	}
	for _, r := range string(bs) {
		if r < 32 || r > 126 {
			return "0x" + hex.EncodeToString(bs)
		}
	}
	return fmt.Sprintf(`%q`, string(bs))
}
