package thirdparty_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"github.com/justenwalker/mack/crypt/random"
	"github.com/justenwalker/mack/internal/testhelpers"
	"github.com/justenwalker/mack/macaroon/thirdparty"
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
		setup     func(iss *MockCaveatIDIssuer)
	}{
		{
			name:      "err-issue-caveat-id",
			expectErr: testErr,
			setup: func(iss *MockCaveatIDIssuer) {
				iss.EXPECT().IssueCaveatID(gomock.Any(), gomock.Any()).Return(nil, testErr)
			},
		},
		{
			name: "issue-caveat-id",
			setup: func(iss *MockCaveatIDIssuer) {
				iss.EXPECT().IssueCaveatID(gomock.Any(), gomock.Any()).Return(cID, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			iss := NewMockCaveatIDIssuer(ctrl)
			tt.setup(iss)
			svc, err := thirdparty.NewAttenuator(thirdparty.AttenuatorConfig{
				Location:     location,
				Scheme:       sch,
				CaveatIssuer: iss,
			})
			if err != nil {
				t.Fatalf("NewDischarger: unexpected error: %v", err)
			}
			seed := int64(1001)
			random.SeedRandom(seed)
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
			if !random.IsTest() {
				// We won't be able to create an identical macaroon, due to nonce differences
				t.Logf("NOTE: Skipping macaroon equality test. Use -tags test_random to run the rest of this test")
				return
			}
			// Attenuate generates the key, so let's generate it again with the same seed
			// which will yield an identical cK and puts the rng in the same state for AddThirdPartyCaveat
			random.SeedRandom(seed)
			cKey := make([]byte, sch.KeySize())
			random.Read(cKey)
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
