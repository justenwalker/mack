package thirdparty_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	macaroon "github.com/justenwalker/mack"
	"github.com/justenwalker/mack/internal/testhelpers"
	"github.com/justenwalker/mack/thirdparty"
)

func TestSet_Discharge(t *testing.T) {
	fx := testhelpers.CreateTestFixture(t, testhelpers.FixtureConfig{
		ID:       "root-id",
		Location: "1st-party",
		Caveats: []testhelpers.Caveat{
			{
				ID:         "third-party-1",
				ThirdParty: "3rd-party",
				Caveats: []testhelpers.Caveat{
					{
						ID:         "third-party-2",
						ThirdParty: "3rd-party-2",
					},
				},
			},
			{
				ID:         "third-party-3",
				ThirdParty: "3rd-party",
				Caveats: []testhelpers.Caveat{
					{
						ID:         "third-party-4",
						ThirdParty: "3rd-party-2",
					},
				},
			},
		},
	})
	var tp1 ThirdPartyMock
	tp1.MatchCaveatFunc = func(c *macaroon.Caveat) bool {
		switch c.Location() {
		case "3rd-party":
			return true
		case "3rd-party-2":
			return false
		default:
			panic(fmt.Errorf("unexpected location: %s", c.Location()))
		}
	}
	tp1.DischargeCaveatFunc = func(_ context.Context, c *macaroon.Caveat) (macaroon.Macaroon, error) {
		switch string(c.ID()) {
		case "third-party-1":
			return fx.Discharge[0], nil
		case "third-party-3":
			return fx.Discharge[2], nil
		default:
			panic(fmt.Errorf("unexpected location: %s", c.Location()))
		}
	}

	var tp2 ThirdPartyMock
	tp2.MatchCaveatFunc = func(c *macaroon.Caveat) bool {
		switch c.Location() {
		case "3rd-party":
			return false
		case "3rd-party-2":
			return true
		default:
			panic(fmt.Errorf("unexpected location: %s", c.Location()))
		}
	}
	tp2.DischargeCaveatFunc = func(_ context.Context, c *macaroon.Caveat) (macaroon.Macaroon, error) {
		switch string(c.ID()) {
		case "third-party-2":
			return fx.Discharge[1], nil
		case "third-party-4":
			return fx.Discharge[3], nil
		default:
			panic(fmt.Errorf("unexpected location: %s", c.Location()))
		}
	}
	set := thirdparty.Set{&tp1, &tp2}
	dm, err := set.Discharge(context.Background(), fx.Target)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(dm) != len(fx.Discharge) {
		t.Fatalf("length mismatch: want %d, got %d", len(dm), len(fx.Discharge))
	}
	sort.Slice(dm, func(i, j int) bool {
		a, b := dm[i], dm[j]
		return bytes.Compare(a.ID(), b.ID()) < 0
	})
	for i := range dm {
		if !fx.Discharge[i].Equal(&dm[i]) {
			t.Errorf("cID is not expected: want %s, got %s", string(fx.Discharge[i].ID()), string(dm[i].ID()))
		}
	}
}

func TestThirdPartyExchange(t *testing.T) {
	ctx := context.Background()
	sch := testhelpers.NewScheme(t)
	m, err := sch.UnsafeRootMacaroon("attenuate", []byte("hello"), testhelpers.RootKey)
	if err != nil {
		t.Fatalf("UnsafeRootMacaroon failed: %v", err)
	}
	m, _ = sch.AddFirstPartyCaveat(&m, []byte("a > 1"))
	m, _ = sch.AddFirstPartyCaveat(&m, []byte("b > 2"))
	tp := newTestService(t, "3p")
	var pcheck PredicateCheckerMock
	dsvc := newTestDischarger(t, "3p")
	tpPredicate := []byte("user = foo")
	m, err = tp.Attenuate(ctx, &m, tpPredicate)
	if err != nil {
		t.Fatalf("Attenuate failed: %v", err)
	}
	tpc := m.ThirdPartyCaveats()
	if len(tpc) < 1 {
		t.Fatalf("expected a third party caveat, got %d", len(m.ThirdPartyCaveats()))
	}
	pcheck.CheckPredicateFunc = func(_ context.Context, predicate []byte) (bool, error) {
		switch string(predicate) {
		case string(tpPredicate):
			return true, nil
		case "a > 1":
			return true, nil
		case "b > 2":
			return true, nil
		default:
			panic(fmt.Errorf("unexpected predicate: %s", string(predicate)))
		}
	}
	dm, err := dsvc.Discharge(ctx, tpc[0].ID(), &pcheck)
	if err != nil {
		t.Fatalf("Discharge failed: %v", err)
	}
	stack, err := sch.PrepareStack(&m, []macaroon.Macaroon{dm})
	if err != nil {
		t.Fatalf("PrepareStack failed: %v", err)
	}
	v, err := sch.Verify(ctx, testhelpers.RootKey, stack)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	err = v.Clear(ctx, &pcheck)
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
}

func newTestService(tb testing.TB, location string) *thirdparty.Attenuator {
	tb.Helper()
	if location == "" {
		location = "3p"
	}
	tp, err := thirdparty.NewAttenuator(thirdparty.AttenuatorConfig{
		Location:     location,
		Scheme:       testhelpers.NewScheme(tb),
		CaveatIssuer: testCaveatIssuer{TB: tb},
	})
	if err != nil {
		tb.Fatalf("unable to create Attenuator: %s", err)
	}
	return tp
}

func newTestDischarger(tb testing.TB, location string) *thirdparty.Discharger {
	tb.Helper()
	if location == "" {
		location = "3p"
	}
	d, err := thirdparty.NewDischarger(thirdparty.DischargerConfig{
		Location:        location,
		Scheme:          testhelpers.NewScheme(tb),
		TicketExtractor: testCaveatIssuer{TB: tb},
	})
	if err != nil {
		tb.Fatalf("unable to create discharger: %s", err)
	}
	return d
}

type testCaveatIssuer struct {
	TB testing.TB
}

type jsonTicket struct {
	CaveatKey []byte `json:"k"`
	Predicate []byte `json:"p"`
}

func (t testCaveatIssuer) IssueCaveatID(_ context.Context, ticket thirdparty.Ticket) ([]byte, error) {
	return json.Marshal(jsonTicket{
		CaveatKey: ticket.CaveatKey,
		Predicate: ticket.Predicate,
	})
}

func (t testCaveatIssuer) ExtractTicket(_ context.Context, cID []byte) (*thirdparty.Ticket, error) {
	var js jsonTicket
	if err := json.Unmarshal(cID, &js); err != nil {
		return nil, err
	}
	return &thirdparty.Ticket{
		CaveatKey: js.CaveatKey,
		Predicate: js.Predicate,
	}, nil
}

var (
	_ thirdparty.CaveatIDIssuer  = testCaveatIssuer{}
	_ thirdparty.TicketExtractor = testCaveatIssuer{}
)
