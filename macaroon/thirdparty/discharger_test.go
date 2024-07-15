package thirdparty_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"github.com/justenwalker/mack/internal/testhelpers"
	"github.com/justenwalker/mack/macaroon"
	"github.com/justenwalker/mack/macaroon/thirdparty"
)

func TestNewDischarger(t *testing.T) {
	tests := []struct {
		name      string
		cfg       thirdparty.DischargerConfig
		expectErr bool
	}{
		{
			name: "success",
			cfg: thirdparty.DischargerConfig{
				Location:        "https://www.example.com",
				Scheme:          &macaroon.Scheme{},
				TicketExtractor: dischargeTestStub{},
			},
		},
		{
			name: "err-no-location",
			cfg: thirdparty.DischargerConfig{
				Location:        "",
				Scheme:          &macaroon.Scheme{},
				TicketExtractor: dischargeTestStub{},
			},
			expectErr: true,
		},
		{
			name: "err-no-Scheme",
			cfg: thirdparty.DischargerConfig{
				Location: "https://www.example.com",
				// Scheme:           &macaroon.Scheme{},
				TicketExtractor: dischargeTestStub{},
			},
			expectErr: true,
		},
		{
			name: "err-no-TicketExtractor",
			cfg: thirdparty.DischargerConfig{
				Location: "https://www.example.com",
				Scheme:   &macaroon.Scheme{},
				// TicketExtractor:  dischargeTestStub{},
			},
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := thirdparty.NewDischarger(tt.cfg)
			if tt.expectErr && err == nil {
				t.Fatal("expected error, but none occurred")
			}
			if err != nil && !tt.expectErr {
				t.Fatal("unexpected error:", err)
			}
		})
	}
}

func TestDischarger_Discharge(t *testing.T) {
	sch := testhelpers.NewScheme(t)
	ctx := context.Background()
	cID := []byte(`caveat-identifier`)
	loc := "https://www.example.com"
	ticket := thirdparty.Ticket{
		CaveatKey: []byte(`12345678901234567890123456789012`),
		Predicate: []byte(`user == foo`),
	}
	m, err := sch.UnsafeRootMacaroon(loc, cID, ticket.CaveatKey)
	if err != nil {
		t.Fatalf("UnsafeRootMacaroon: unexpected error: %v", err)
	}
	testErr := errors.New("test error")

	tests := []struct {
		name      string
		expectErr error
		expected  *macaroon.Macaroon
		setup     func(text *MockTicketExtractor, pcheck *MockPredicateChecker)
	}{
		{
			name:      "err-extract-ticket",
			expectErr: testErr,
			setup: func(text *MockTicketExtractor, _ *MockPredicateChecker) {
				text.EXPECT().ExtractTicket(ctx, cID).Return(nil, testErr)
			},
		},
		{
			name:      "err-check-predicate",
			expectErr: testErr,
			setup: func(text *MockTicketExtractor, pcheck *MockPredicateChecker) {
				text.EXPECT().ExtractTicket(ctx, cID).Return(&ticket, nil)
				pcheck.EXPECT().CheckPredicate(gomock.Any(), gomock.Eq(matchBytes(ticket.Predicate))).Return(false, testErr)
			},
		},
		{
			name:      "fail-check-predicate",
			expectErr: macaroon.ErrPredicateNotSatisfied,
			setup: func(text *MockTicketExtractor, pcheck *MockPredicateChecker) {
				text.EXPECT().ExtractTicket(ctx, cID).Return(&ticket, nil)
				pcheck.EXPECT().CheckPredicate(gomock.Any(), gomock.Eq(matchBytes(ticket.Predicate))).Return(false, nil)
			},
		},
		{
			name: "success",
			setup: func(text *MockTicketExtractor, pcheck *MockPredicateChecker) {
				text.EXPECT().ExtractTicket(ctx, cID).Return(&ticket, nil)
				pcheck.EXPECT().CheckPredicate(gomock.Any(), gomock.Eq(matchBytes(ticket.Predicate))).Return(true, nil)
			},
			expected: &m,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			text := NewMockTicketExtractor(ctrl)
			pcheck := NewMockPredicateChecker(ctrl)
			tt.setup(text, pcheck)
			ds, err := thirdparty.NewDischarger(thirdparty.DischargerConfig{
				Location:        loc,
				Scheme:          sch,
				TicketExtractor: text,
			})
			if err != nil {
				t.Fatalf("NewDischarger: unexpected error: %v", err)
			}
			am, err := ds.Discharge(ctx, cID, pcheck)
			if tt.expectErr != nil {
				if !errors.Is(err, tt.expectErr) {
					t.Fatalf("want %v, got %v", tt.expectErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.expected, &am, cmp.Comparer(macaroonEqual)); diff != "" {
				t.Fatalf("Discharge unexpected macaroon (-want +got):\n%s", diff)
			}
		})
	}
}

func macaroonEqual(a, b macaroon.Macaroon) bool {
	return a.Equal(&b)
}

type dischargeTestStub struct{}

func (d dischargeTestStub) CheckPredicate(_ context.Context, _ []byte) (bool, error) {
	panic("unimplemented")
}

func (d dischargeTestStub) ExtractTicket(_ context.Context, _ []byte) (*thirdparty.Ticket, error) {
	panic("unimplemented")
}

var (
	_ thirdparty.TicketExtractor  = dischargeTestStub{}
	_ thirdparty.PredicateChecker = dischargeTestStub{}
)