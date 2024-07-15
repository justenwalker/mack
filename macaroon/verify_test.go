package macaroon_test

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/justenwalker/mack/internal/testhelpers"
	"github.com/justenwalker/mack/macaroon"
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
		setup     func(t *testing.T, m *testhelpers.MockPredicateChecker) testhelpers.Fixture
		success   bool
		expectErr func(t testing.TB, err error)
	}{
		{
			name: "success",
			setup: func(t *testing.T, m *testhelpers.MockPredicateChecker) testhelpers.Fixture {
				t.Helper()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`a > 1`)).Return(true, nil).AnyTimes()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`b > 2`)).Return(true, nil).AnyTimes()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`user = foo`)).Return(true, nil).AnyTimes()
				return testhelpers.CreateTestFixture(t, cfg)
			},
			success: true,
		},
		{
			name: "caveat-fail",
			setup: func(t *testing.T, m *testhelpers.MockPredicateChecker) testhelpers.Fixture {
				t.Helper()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`a > 1`)).Return(true, nil).AnyTimes()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`b > 2`)).Return(false, nil).AnyTimes()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`user = foo`)).Return(true, nil).AnyTimes()
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
			setup: func(t *testing.T, m *testhelpers.MockPredicateChecker) testhelpers.Fixture {
				t.Helper()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`a > 1`)).Return(true, nil).AnyTimes()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`b > 2`)).Return(false, errors.New("failed")).AnyTimes()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`user = foo`)).Return(true, nil).AnyTimes()
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
			setup: func(t *testing.T, m *testhelpers.MockPredicateChecker) testhelpers.Fixture {
				t.Helper()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`a > 1`)).Return(true, nil).AnyTimes()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`b > 2`)).Return(true, nil).AnyTimes()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`user = foo`)).Return(true, nil).AnyTimes()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`c > 3`)).Return(true, nil).AnyTimes()
				return testhelpers.CreateTestFixture(t, nestedCfg)
			},
			success: true,
		},
		{
			name: "nested-caveat-fail",
			setup: func(t *testing.T, m *testhelpers.MockPredicateChecker) testhelpers.Fixture {
				t.Helper()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`a > 1`)).Return(true, nil).AnyTimes()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`b > 2`)).Return(true, nil).AnyTimes()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`user = foo`)).Return(true, nil).AnyTimes()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`c > 3`)).Return(false, nil).AnyTimes()
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
			setup: func(t *testing.T, m *testhelpers.MockPredicateChecker) testhelpers.Fixture {
				t.Helper()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`a > 1`)).Return(true, nil).AnyTimes()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`b > 2`)).Return(true, nil).AnyTimes()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`user = foo`)).Return(true, nil).AnyTimes()
				m.EXPECT().CheckPredicate(gomock.Any(), []byte(`c > 3`)).Return(false, errors.New("failed")).AnyTimes()
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
			ctrl := gomock.NewController(t)
			m := testhelpers.NewMockPredicateChecker(ctrl)
			fx := tt.setup(t, m)
			v := macaroon.InsecureVerifiedStack(fx.Stack)
			err := v.Clear(ctx, m)
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
