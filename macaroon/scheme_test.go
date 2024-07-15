package macaroon_test

import (
	"context"
	"errors"
	"testing"

	"github.com/justenwalker/mack/internal/testhelpers"
	"github.com/justenwalker/mack/macaroon"
)

func TestScheme_Validate(t *testing.T) {
	ctx := context.Background()
	cfg := testhelpers.FixtureConfig{
		ID: "hello",
		Caveats: []testhelpers.Caveat{
			{
				ID: "a > 1",
			},
			{
				ID: "b > 2",
			},
			{
				ID:         "{cK,userid == foo}",
				ThirdParty: "https://other.example.org",
			},
			{
				ID: "user = foo",
			},
		},
	}
	t.Run("bad-key", func(t *testing.T) {
		fx := testhelpers.CreateTestFixture(t, cfg)
		_, err := fx.Scheme.Verify(ctx, []byte(`foo`), fx.Stack)
		if !errors.Is(err, macaroon.ErrInvalidArgument) {
			t.Fatalf("expected macaroon.ErrInvalidArgument but was %TB", err)
		}
	})
	t.Run("verify", func(t *testing.T) {
		fx := testhelpers.CreateTestFixture(t, cfg)
		_, err := fx.Scheme.Verify(ctx, testhelpers.RootKey, fx.Stack)
		if err != nil {
			t.Logf("m: %v", fx.Stack.Target())
			for _, d := range fx.Stack.Discharges() {
				t.Logf("d: %v", &d) //nolint:gosec
			}
			t.Fatalf("Clear: %v", errors.Unwrap(err))
		}
	})
	t.Run("not-bound", func(t *testing.T) {
		fx := testhelpers.CreateTestFixture(t, cfg)
		stack := make([]macaroon.Macaroon, 0, len(fx.Stack))
		stack = append(stack, *fx.Target)
		stack = append(stack, fx.Discharge...)
		_, err := fx.Scheme.Verify(ctx, testhelpers.RootKey, stack)
		if !errors.Is(err, macaroon.ErrVerificationFailed) {
			t.Fatalf("expected macaroon.ErrVerificationFailed: was %TB", err)
		}
		t.Logf("validation failed: %v <%[1]TB>", errors.Unwrap(err))
	})
	t.Run("bad-key", func(t *testing.T) {
		badKey := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7}
		fx := testhelpers.CreateTestFixture(t, cfg)
		_, err := fx.Scheme.Verify(ctx, badKey, fx.Stack)
		if !errors.Is(err, macaroon.ErrVerificationFailed) {
			t.Fatalf("expected macaroon.ErrVerificationFailed: was %TB", err)
		}
		t.Logf("validation failed: %v <%[1]TB>", errors.Unwrap(err))
	})
	t.Run("no-discharge", func(t *testing.T) {
		fx := testhelpers.CreateTestFixture(t, cfg)
		_, err := fx.Scheme.Verify(ctx, testhelpers.RootKey, macaroon.Stack([]macaroon.Macaroon{*fx.Target}))
		if !errors.Is(err, macaroon.ErrVerificationFailed) {
			t.Fatalf("expected macaroon.ErrVerificationFailed: was %TB", err)
		}
		t.Logf("validation failed: %v <%[1]TB>", errors.Unwrap(err))
	})
}
