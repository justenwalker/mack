package thirdparty_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/justenwalker/mack/internal/testhelpers"
	"github.com/justenwalker/mack/macaroon"
	"github.com/justenwalker/mack/macaroon/thirdparty"
)

//go:generate go run go.uber.org/mock/mockgen -source thirdparty.go -destination thirdparty_mock_test.go -package thirdparty_test

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
	ctrl := gomock.NewController(t)
	tp1 := NewMockThirdParty(ctrl)
	tp1.EXPECT().MatchCaveat(matchLocation("3rd-party")).Return(true).AnyTimes()
	tp1.EXPECT().MatchCaveat(matchLocation("3rd-party-2")).Return(false).AnyTimes()
	tp1.EXPECT().DischargeCaveat(gomock.Any(), matchCaveatID("third-party-1")).Return(fx.Discharge[0], nil)
	tp1.EXPECT().DischargeCaveat(gomock.Any(), matchCaveatID("third-party-3")).Return(fx.Discharge[2], nil)
	tp2 := NewMockThirdParty(ctrl)
	tp2.EXPECT().MatchCaveat(matchLocation("3rd-party")).Return(false).AnyTimes()
	tp2.EXPECT().MatchCaveat(matchLocation("3rd-party-2")).Return(true).AnyTimes()
	tp2.EXPECT().DischargeCaveat(gomock.Any(), matchCaveatID("third-party-2")).Return(fx.Discharge[1], nil)
	tp2.EXPECT().DischargeCaveat(gomock.Any(), matchCaveatID("third-party-4")).Return(fx.Discharge[3], nil)
	set := thirdparty.Set{tp1, tp2}
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

type matchLocation string

func (m matchLocation) Matches(x any) bool {
	c, ok := x.(*macaroon.Caveat)
	if !ok {
		return false
	}
	loc := c.Location()
	return loc == string(m)
}

func (m matchLocation) String() string {
	return "caveat-loc=" + string(m)
}

type matchCaveatID string

func (m matchCaveatID) Matches(x any) bool {
	c, ok := x.(*macaroon.Caveat)
	if !ok {
		return false
	}
	id := string(c.ID())
	return id == string(m)
}

func (m matchCaveatID) String() string {
	return "caveat-id=" + string(m)
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
	if err != nil {
		t.Fatalf("UnsafeRootMacaroon failed: %v", err)
	}
	tp := newTestService(t, "3p")
	ctrl := gomock.NewController(t)
	pcheck := NewMockPredicateChecker(ctrl)
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
	pcheck.EXPECT().CheckPredicate(gomock.Any(), tpPredicate).Return(true, nil)
	pcheck.EXPECT().CheckPredicate(gomock.Any(), []byte("a > 1")).Return(true, nil)
	pcheck.EXPECT().CheckPredicate(gomock.Any(), []byte("b > 2")).Return(true, nil)

	dm, err := dsvc.Discharge(ctx, tpc[0].ID(), pcheck)
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
	err = v.Clear(ctx, pcheck)
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

type matchBytes []byte

func (mb matchBytes) Matches(x any) bool {
	if bs, ok := x.([]byte); ok {
		return bytesEqual(mb, bs)
	}
	return false
}

func (mb matchBytes) String() string {
	return fmt.Sprintf("[]byte(%x)", []byte(mb))
}

var _ gomock.Matcher = &matchBytes{}

func bytesEqual(a []byte, b []byte) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	return bytes.Equal(a, b)
}
