package macaroon_test

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/justenwalker/mack/crypt/random"
	"github.com/justenwalker/mack/internal/testhelpers"
	"github.com/justenwalker/mack/macaroon"
)

func TestTraces_fail(t *testing.T) {
	ctx := context.Background()
	ctx = macaroon.WithVerifyContext(ctx)
	sch := testhelpers.NewScheme(t)
	random.SeedRandom(1000)
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
	if !random.IsTest() {
		// We won't be able to inspect trace for exact match, due to nonce differences
		t.Logf("NOTE: Skipping trace equality test. Use -tags test_random to run the rest of this test")
		return
	}
	expected := macaroon.Traces{
		{
			RootKey: testhelpers.RootKey,
			Ops: []*macaroon.TraceOp{
				{
					Kind:   macaroon.TraceOpHMAC,
					Arg1:   testhelpers.RootKey,
					Arg2:   []byte(cfg.ID),
					Result: hexBytes("0xbc3d4ca4dc1e239400c01c7a7955f955aa42c2082b2f6c5b716bd98c1b62c1b2"),
				},
				{
					Kind:   macaroon.TraceOpHMAC,
					Arg1:   hexBytes("0xbc3d4ca4dc1e239400c01c7a7955f955aa42c2082b2f6c5b716bd98c1b62c1b2"),
					Arg2:   []byte(cfg.Caveats[0].ID),
					Result: hexBytes("0x4097bd90962e4e9f6f6f97650cc848a5c0c6d98f083db15b0c960e6ef8f4673a"),
				},
				{
					Kind:   macaroon.TraceOpHMAC,
					Arg1:   hexBytes("0x4097bd90962e4e9f6f6f97650cc848a5c0c6d98f083db15b0c960e6ef8f4673a"),
					Arg2:   []byte(cfg.Caveats[1].ID),
					Result: hexBytes("0xd588df38d0bcc8d15fa8e282b41163b083623878a2d5164f433327f833c871a6"),
				},
				{
					Kind:   macaroon.TraceOpDecrypt,
					Arg1:   hexBytes("0xd588df38d0bcc8d15fa8e282b41163b083623878a2d5164f433327f833c871a6"),
					Arg2:   hexBytes("0xbebce5518d6fdc1167b07e39df9300231ab435fab838e7898d0c2bf8d9a40a747f8a56d8b5fb1f038ffbfce79f185f4aad9d603094edb854"),
					Result: testhelpers.ThirdPartyKey,
				},
				{
					Kind:  macaroon.TraceOpFail,
					Error: macaroon.ErrVerificationFailed,
				},
			},
		},
		{
			RootKey: testhelpers.ThirdPartyKey,
			Ops: []*macaroon.TraceOp{
				{
					Kind:   macaroon.TraceOpHMAC,
					Arg1:   testhelpers.ThirdPartyKey,
					Arg2:   []byte(cfg.Caveats[2].ID),
					Result: hexBytes("0x4ec371d54f699dc7f8ac3f58ca5db70d4da959fec1032e38aaf674a40220d1e6"),
				},
				{
					Kind:   macaroon.TraceOpBind,
					Arg1:   hexBytes("0x8ac92fd7d1f4dbdf0655e0939bffd827e995e9de4cd29fa304b6fb060fe9e61c"),
					Arg2:   hexBytes("0x4ec371d54f699dc7f8ac3f58ca5db70d4da959fec1032e38aaf674a40220d1e6"),
					Result: hexBytes("0x69e024928f4373cb5f2ce6a2188711a1267152f226b8f17a95fd7c7b4e055f04"),
				},
			},
		},
	}
	if diff := cmp.Diff(expected, traces, cmp.Comparer(traceErrorEquals)); diff != "" {
		t.Fatalf("traces mismatched (-want +got):\n%s", diff)
	}
}

func TestTraces_success(t *testing.T) {
	ctx := context.Background()
	ctx = macaroon.WithVerifyContext(ctx)
	sch := testhelpers.NewScheme(t)
	random.SeedRandom(1000)
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
	if !random.IsTest() {
		// We won't be able to inspect trace for exact match, due to nonce differences
		t.Logf("NOTE: Skipping trace equality test. Use -tags test_random to run the rest of this test")
		return
	}
	expected := macaroon.Traces{
		{
			RootKey: testhelpers.RootKey,
			Ops: []*macaroon.TraceOp{
				{
					Kind:   macaroon.TraceOpHMAC,
					Arg1:   testhelpers.RootKey,
					Arg2:   []byte(cfg.ID),
					Result: hexBytes("0xbc3d4ca4dc1e239400c01c7a7955f955aa42c2082b2f6c5b716bd98c1b62c1b2"),
				},
				{
					Kind:   macaroon.TraceOpHMAC,
					Arg1:   hexBytes("0xbc3d4ca4dc1e239400c01c7a7955f955aa42c2082b2f6c5b716bd98c1b62c1b2"),
					Arg2:   []byte(cfg.Caveats[0].ID),
					Result: hexBytes("0x4097bd90962e4e9f6f6f97650cc848a5c0c6d98f083db15b0c960e6ef8f4673a"),
				},
				{
					Kind:   macaroon.TraceOpHMAC,
					Arg1:   hexBytes("0x4097bd90962e4e9f6f6f97650cc848a5c0c6d98f083db15b0c960e6ef8f4673a"),
					Arg2:   []byte(cfg.Caveats[1].ID),
					Result: hexBytes("0xd588df38d0bcc8d15fa8e282b41163b083623878a2d5164f433327f833c871a6"),
				},
				{
					Kind:   macaroon.TraceOpDecrypt,
					Arg1:   hexBytes("0xd588df38d0bcc8d15fa8e282b41163b083623878a2d5164f433327f833c871a6"),
					Arg2:   hexBytes("0xbebce5518d6fdc1167b07e39df9300231ab435fab838e7898d0c2bf8d9a40a747f8a56d8b5fb1f038ffbfce79f185f4aad9d603094edb854"),
					Result: testhelpers.ThirdPartyKey,
				},
				{
					Kind:   macaroon.TraceOpHMAC,
					Arg1:   hexBytes("0xd588df38d0bcc8d15fa8e282b41163b083623878a2d5164f433327f833c871a6"),
					Arg2:   hexBytes("0xbebce5518d6fdc1167b07e39df9300231ab435fab838e7898d0c2bf8d9a40a747f8a56d8b5fb1f038ffbfce79f185f4aad9d603094edb8547b634b2c757365726964203d3d20666f6f7d"),
					Result: hexBytes("0x9ed58f98b46b5c39a93174ad6c5bac7b6a7cf6ade79444f90472e30db8a8cabb"),
				},
				{
					Kind:   macaroon.TraceOpHMAC,
					Arg1:   hexBytes("0x9ed58f98b46b5c39a93174ad6c5bac7b6a7cf6ade79444f90472e30db8a8cabb"),
					Arg2:   []byte(cfg.Caveats[3].ID),
					Result: hexBytes("0x8ac92fd7d1f4dbdf0655e0939bffd827e995e9de4cd29fa304b6fb060fe9e61c"),
				},
			},
		},
		{
			RootKey: testhelpers.ThirdPartyKey,
			Ops: []*macaroon.TraceOp{
				{
					Kind:   macaroon.TraceOpHMAC,
					Arg1:   testhelpers.ThirdPartyKey,
					Arg2:   []byte(cfg.Caveats[2].ID),
					Result: hexBytes("0x4ec371d54f699dc7f8ac3f58ca5db70d4da959fec1032e38aaf674a40220d1e6"),
				},
				{
					Kind:   macaroon.TraceOpBind,
					Arg1:   hexBytes("0x8ac92fd7d1f4dbdf0655e0939bffd827e995e9de4cd29fa304b6fb060fe9e61c"),
					Arg2:   hexBytes("0x4ec371d54f699dc7f8ac3f58ca5db70d4da959fec1032e38aaf674a40220d1e6"),
					Result: hexBytes("0x69e024928f4373cb5f2ce6a2188711a1267152f226b8f17a95fd7c7b4e055f04"),
				},
			},
		},
	}
	if diff := cmp.Diff(expected, traces); diff != "" {
		t.Fatalf("traces mismatched (-want +got):\n%s", diff)
	}
}

func traceErrorEquals(a, b error) bool {
	return errors.Is(b, a)
}

func hexBytes(s string) []byte {
	b, _ := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	return b
}
