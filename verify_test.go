package mack_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	macaroon "github.com/justenwalker/mack"
	"github.com/justenwalker/mack/internal/testhelpers"
)

func TestVerified_Check(t *testing.T) {
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
	nestedCfg := testhelpers.FixtureConfig{
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
				Caveats: []testhelpers.Caveat{
					{
						ID: "c > 3",
					},
				},
			},
			{
				ID: "user = foo",
			},
		},
	}
	tests := []struct {
		name      string
		setup     func(t *testing.T, m *PredicateCheckerMock) testhelpers.Fixture
		success   bool
		expectErr func(t testing.TB, err error)
	}{
		{
			name: "success",
			setup: func(t *testing.T, m *PredicateCheckerMock) testhelpers.Fixture {
				t.Helper()
				m.CheckPredicateFunc = func(_ context.Context, p []byte) (bool, error) {
					if bytes.Equal(p, []byte(`a > 1`)) {
						return true, nil
					}
					if bytes.Equal(p, []byte(`b > 2`)) {
						return true, nil
					}
					if bytes.Equal(p, []byte(`user = foo`)) {
						return true, nil
					}
					return false, fmt.Errorf("unexpected predicate: %s", string(p))
				}
				return testhelpers.CreateTestFixture(t, cfg)
			},
			success: true,
		},
		{
			name: "caveat-fail",
			setup: func(t *testing.T, m *PredicateCheckerMock) testhelpers.Fixture {
				t.Helper()
				m.CheckPredicateFunc = func(_ context.Context, p []byte) (bool, error) {
					if bytes.Equal(p, []byte(`a > 1`)) {
						return true, nil
					}
					if bytes.Equal(p, []byte(`b > 2`)) {
						return false, nil
					}
					if bytes.Equal(p, []byte(`user = foo`)) {
						return true, nil
					}
					return false, fmt.Errorf("unexpected predicate: %s", string(p))
				}
				return testhelpers.CreateTestFixture(t, cfg)
			},
			success: false,
			expectErr: func(tb testing.TB, err error) {
				tb.Helper()
				if !errors.Is(err, macaroon.ErrPredicateNotSatisfied) {
					tb.Errorf("expected errors.Is(err, macaroon.ErrPredicateNotSatisfied): Error %#v <%[1]TB>", err)
				}
			},
		},
		{
			name: "caveat-error",
			setup: func(t *testing.T, m *PredicateCheckerMock) testhelpers.Fixture {
				t.Helper()
				m.CheckPredicateFunc = func(_ context.Context, p []byte) (bool, error) {
					if bytes.Equal(p, []byte(`a > 1`)) {
						return true, nil
					}
					if bytes.Equal(p, []byte(`b > 2`)) {
						return false, errors.New("failed")
					}
					if bytes.Equal(p, []byte(`user = foo`)) {
						return true, nil
					}
					return false, fmt.Errorf("unexpected predicate: %s", string(p))
				}
				return testhelpers.CreateTestFixture(t, cfg)
			},
			success: false,
			expectErr: func(tb testing.TB, err error) {
				tb.Helper()
				if errors.Is(err, macaroon.ErrPredicateNotSatisfied) {
					tb.Errorf("expected !errors.Is(err, macaroon.ErrPredicateNotSatisfied): Error %#v <%[1]TB>", err)
				}
			},
		},
		{
			name: "nested-success",
			setup: func(t *testing.T, m *PredicateCheckerMock) testhelpers.Fixture {
				t.Helper()
				m.CheckPredicateFunc = func(_ context.Context, p []byte) (bool, error) {
					if bytes.Equal(p, []byte(`a > 1`)) {
						return true, nil
					}
					if bytes.Equal(p, []byte(`b > 2`)) {
						return true, nil
					}
					if bytes.Equal(p, []byte(`user = foo`)) {
						return true, nil
					}
					if bytes.Equal(p, []byte(`c > 3`)) {
						return true, nil
					}
					return false, fmt.Errorf("unexpected predicate: %s", string(p))
				}
				return testhelpers.CreateTestFixture(t, nestedCfg)
			},
			success: true,
		},
		{
			name: "nested-caveat-fail",
			setup: func(t *testing.T, m *PredicateCheckerMock) testhelpers.Fixture {
				t.Helper()
				m.CheckPredicateFunc = func(_ context.Context, p []byte) (bool, error) {
					if bytes.Equal(p, []byte(`a > 1`)) {
						return true, nil
					}
					if bytes.Equal(p, []byte(`b > 2`)) {
						return true, nil
					}
					if bytes.Equal(p, []byte(`user = foo`)) {
						return true, nil
					}
					if bytes.Equal(p, []byte(`c > 3`)) {
						return false, nil
					}
					return false, fmt.Errorf("unexpected predicate: %s", string(p))
				}
				return testhelpers.CreateTestFixture(t, nestedCfg)
			},
			success: false,
			expectErr: func(tb testing.TB, err error) {
				tb.Helper()
				if !errors.Is(err, macaroon.ErrPredicateNotSatisfied) {
					tb.Errorf("expected errors.Is(err, macaroon.ErrPredicateNotSatisfied): Error %#v <%[1]TB>", err)
				}
			},
		},
		{
			name: "nested-caveat-error",
			setup: func(t *testing.T, m *PredicateCheckerMock) testhelpers.Fixture {
				t.Helper()
				m.CheckPredicateFunc = func(_ context.Context, p []byte) (bool, error) {
					if bytes.Equal(p, []byte(`a > 1`)) {
						return true, nil
					}
					if bytes.Equal(p, []byte(`b > 2`)) {
						return true, nil
					}
					if bytes.Equal(p, []byte(`user = foo`)) {
						return true, nil
					}
					if bytes.Equal(p, []byte(`c > 3`)) {
						return false, errors.New("failed")
					}
					return false, fmt.Errorf("unexpected predicate: %s", string(p))
				}
				return testhelpers.CreateTestFixture(t, nestedCfg)
			},
			success: false,
			expectErr: func(tb testing.TB, err error) {
				tb.Helper()
				if errors.Is(err, macaroon.ErrPredicateNotSatisfied) {
					tb.Errorf("expected !errors.Is(err, macaroon.ErrPredicateNotSatisfied): Error %#v <%[1]TB>", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var m PredicateCheckerMock
			fx := tt.setup(t, &m)
			v := macaroon.InsecureVerifiedStack(fx.Stack)
			err := v.Clear(ctx, &m)
			if err != nil {
				if tt.success {
					t.Fatalf("unexpected verify error: %v", err)
				}
				t.Logf("verify error: %v", err)
				if tt.expectErr != nil {
					tt.expectErr(t, err)
				}
				return
			}
			if !tt.success {
				t.Fatalf("expected verify error, but got none")
			}
		})
	}
}
