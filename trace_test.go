package mack_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sebdah/goldie/v2"

	macaroon "github.com/justenwalker/mack"
	"github.com/justenwalker/mack/internal/testhelpers"
)

func TestTraces_fail(t *testing.T) {
	g := goldie.New(t, goldie.WithFixtureDir("testdata/traces"))
	ctx := context.Background()
	ctx = macaroon.WithVerifyContext(ctx)
	sch := testhelpers.NewScheme(t)
	testhelpers.SeedRandom(1000)
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
	fx := testhelpers.CreateTestFixture(t, cfg)
	stack := make([]macaroon.Macaroon, 0, len(fx.Stack))
	stack = append(stack, *fx.Target)
	stack = append(stack, fx.Discharge...) // not bound to request
	_, err := sch.Verify(ctx, testhelpers.RootKey, stack)
	if err == nil {
		t.Fatalf("expected verify to fail")
	}
	traces := macaroon.GetTraces(ctx)
	t.Log(traces.String())
	if len(traces) != 2 {
		t.Fatalf("expected 2 traces, got %d", len(traces))
	}
	if diff := cmp.Diff(testhelpers.RootKey, traces[0].RootKey); diff != "" {
		t.Fatalf("trace[0].RootKey mismatch (-want +got):\n%s", diff)
	}
	if len(traces[0].Ops) != 5 {
		t.Fatalf("expected trace[0] to have 5 operations, got %d", len(traces[0].Ops))
	}
	if diff := cmp.Diff(testhelpers.ThirdPartyKey, traces[1].RootKey); diff != "" {
		t.Fatalf("trace[0].RootKey mismatch (-want +got):\n%s", diff)
	}
	if len(traces[1].Ops) != 2 {
		t.Fatalf("expected trace[1] to have 2 operations, got %d", len(traces[1].Ops))
	}
	g.Assert(t, "TestTraces_fail", []byte(traces.String()))
}

func TestTraces_success(t *testing.T) {
	g := goldie.New(t, goldie.WithFixtureDir("testdata/traces"))
	ctx := context.Background()
	ctx = macaroon.WithVerifyContext(ctx)
	sch := testhelpers.NewScheme(t)
	testhelpers.SeedRandom(1000)
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
	fx := testhelpers.CreateTestFixture(t, cfg)
	_, err := sch.Verify(ctx, testhelpers.RootKey, fx.Stack)
	if err != nil {
		t.Fatalf("failed to verify traces: %s", err)
	}
	traces := macaroon.GetTraces(ctx)
	t.Log(traces.String())
	if len(traces) != 2 {
		t.Fatalf("expected 2 traces, got %d", len(traces))
	}
	if diff := cmp.Diff(testhelpers.RootKey, traces[0].RootKey); diff != "" {
		t.Fatalf("trace[0].RootKey mismatch (-want +got):\n%s", diff)
	}
	if len(traces[0].Ops) != 6 {
		t.Fatalf("expected trace[0] to have 6 operations, got %d", len(traces[0].Ops))
	}
	if diff := cmp.Diff(testhelpers.ThirdPartyKey, traces[1].RootKey); diff != "" {
		t.Fatalf("trace[0].RootKey mismatch (-want +got):\n%s", diff)
	}
	if len(traces[1].Ops) != 2 {
		t.Fatalf("expected trace[1] to have 2 operations, got %d", len(traces[1].Ops))
	}
	g.Assert(t, "TestTraces_success", []byte(traces.String()))
}
