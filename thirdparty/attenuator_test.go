package thirdparty_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/justenwalker/mack/internal/testhelpers"
	"github.com/justenwalker/mack/thirdparty"
)

func TestService_Attenuate(t *testing.T) {
	sch := testhelpers.NewScheme(t)
	ctx := context.Background()
	location := "3p"
	m, err := sch.UnsafeRootMacaroon("1p", []byte(`hello`), testhelpers.RootKey)
	if err != nil {
		t.Fatalf("UnsafeRootMacaroon: %v", err)
	}
	cID := []byte(`caveat-id`)
	ticket := thirdparty.Ticket{
		CaveatKey: testhelpers.ThirdPartyKey,
		Predicate: []byte(`user == foo`),
	}
	testErr := errors.New("test error")
	tests := []struct {
		name      string
		expectErr error
		setup     func(iss *CaveatIDIssuerMock)
	}{
		{
			name:      "err-issue-caveat-id",
			expectErr: testErr,
			setup: func(iss *CaveatIDIssuerMock) {
				iss.IssueCaveatIDFunc = func(context.Context, thirdparty.Ticket) ([]byte, error) {
					return nil, testErr
				}
			},
		},
		{
			name: "issue-caveat-id",
			setup: func(iss *CaveatIDIssuerMock) {
				iss.IssueCaveatIDFunc = func(context.Context, thirdparty.Ticket) ([]byte, error) {
					return cID, nil
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mock CaveatIDIssuerMock
			tt.setup(&mock)
			svc, err := thirdparty.NewAttenuator(thirdparty.AttenuatorConfig{
				Location:     location,
				Scheme:       sch,
				CaveatIssuer: &mock,
			}, thirdparty.WithRandSource(testhelpers.ReadRandom))
			if err != nil {
				t.Fatalf("NewDischarger: unexpected error: %v", err)
			}
			seed := int64(1001)
			testhelpers.SeedRandom(seed)
			am, err := svc.Attenuate(ctx, &m, ticket.Predicate)
			if tt.expectErr != nil {
				if !errors.Is(err, tt.expectErr) {
					t.Fatalf("want %v, got %v", tt.expectErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			cav := am.Caveats()[0]
			if diff := cmp.Diff(cID, cav.ID()); diff != "" {
				t.Fatalf("caveat.ID: (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(location, cav.Location()); diff != "" {
				t.Fatalf("caveat.Location: (-want +got):\n%s", diff)
			}
			// Attenuate generates the key, so let's generate it again with the same seed
			// which will yield an identical cK and puts the rng in the same state for AddThirdPartyCaveat
			testhelpers.SeedRandom(seed)
			cKey := make([]byte, sch.KeySize())
			testhelpers.MustReadRandom(cKey)
			tm, err := sch.AddThirdPartyCaveat(&m, cKey, cav.ID(), location)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(am, tm, cmp.Comparer(macaroonEqual)); diff != "" {
				t.Fatalf("Macaroons do not match: (-want +got):\n%s", diff)
			}
		})
	}
}
