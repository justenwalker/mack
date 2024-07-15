package macaroon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=TraceOpKind -linecomment -output trace_string.go

type TraceOpKind int

const (
	TraceOpUnknown = TraceOpKind(iota) // Unknown
	TraceOpHMAC                        // HMAC
	TraceOpDecrypt                     // Decrypt
	TraceOpBind                        // BindForRequest
	TraceOpFail                        // FAILURE
)

// TraceOp represents an operation performed on a macaroon that is recorded in the trace.
type TraceOp struct {
	Kind   TraceOpKind
	Arg1   []byte
	Arg2   []byte
	Result []byte
	Error  error
}

type jsonTraceOp struct {
	Kind   string   `json:"kind,omitempty"`
	Args   []string `json:"args,omitempty"`
	Result string   `json:"result,omitempty"`
	Error  []string `json:"error,omitempty"`
}

func (op *TraceOp) MarshalJSON() ([]byte, error) {
	args := make([]string, 0, 2)
	var traceErrors []string
	err := op.Error
	for err != nil {
		traceErrors = append(traceErrors, err.Error())
		err = errors.Unwrap(err)
	}
	if op.Arg1 != nil {
		args = append(args, printableBytes(op.Arg1))
	}
	if op.Arg2 != nil {
		args = append(args, printableBytes(op.Arg2))
	}
	return json.Marshal(jsonTraceOp{
		Kind:   op.Kind.String(),
		Args:   args,
		Result: printableBytes(op.Result),
		Error:  traceErrors,
	})
}

// WithVerifyContext creates a new context with a trace structure value attached.
// This value is used by the Verify function to store the performed traces.
// This is useful for debugging why macaroon verification failed.
func WithVerifyContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, verifyContextKey{}, &verifyContext{})
}

// Traces represents a trace of the verification operations done during macaroon verification.
// This structure is returned by [GetTraces] when verifying a macaroon with a context primed with [WithVerifyContext].
type Traces []Trace

type tracesJSON struct {
	Traces []Trace `json:"traces"`
}

func (t Traces) String() string {
	js, err := json.MarshalIndent(tracesJSON{Traces: t}, "", "  ")
	if err != nil {
		return fmt.Sprintf("Traces: FAILED TO RENDER JSON: %v", err)
	}
	return string(js)
}

// GetTraces returns the [Traces] performed during [Scheme.Verify] if a context primed with [WithVerifyContext] was provided.
func GetTraces(ctx context.Context) Traces {
	vc := getVerifyContext(ctx)
	if vc == nil {
		return nil
	}
	return vc.stacks
}

type verifyContext struct {
	stacks []Trace
}

type verifyContextKey struct{}

// Trace contains the operations done on a single macaroon.
type Trace struct {
	RootKey []byte
	Ops     []*TraceOp
}

type jsonTrace struct {
	RootKey string     `json:"rootKey"`
	Ops     []*TraceOp `json:"ops"`
}

func (t *Trace) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonTrace{
		RootKey: printableBytes(t.RootKey),
		Ops:     t.Ops,
	})
}

func (v *verifyContext) init(stack Stack) {
	if v == nil {
		return
	}
	v.stacks = make([]Trace, len(stack))
}

func (v *verifyContext) traceRootKey(index int, key []byte, id []byte) *TraceOp {
	if v == nil {
		return nil
	}
	v.stacks[index].RootKey = cloneBytes(key)
	return v.trace(index, TraceOpHMAC, v.stacks[index].RootKey, cloneBytes(id))
}

func (v *verifyContext) trace(index int, kind TraceOpKind, arg1, arg2 []byte) *TraceOp {
	if v == nil {
		return nil
	}
	op := TraceOp{
		Kind: kind,
		Arg1: cloneBytes(arg1),
		Arg2: cloneBytes(arg2),
	}
	v.stacks[index].Ops = append(v.stacks[index].Ops, &op)
	return &op
}

func (v *verifyContext) fail(index int, err error) {
	v.trace(index, TraceOpFail, nil, nil).setError(err)
}

func (op *TraceOp) setResult(r []byte) {
	if op == nil {
		return
	}
	op.Result = cloneBytes(r)
}

func (op *TraceOp) setError(err error) {
	if op == nil {
		return
	}
	op.Error = err
}

func cloneBytes(vs []byte) []byte {
	if vs == nil {
		return nil
	}
	bs := make([]byte, len(vs))
	copy(bs, vs)
	return bs
}

func getVerifyContext(ctx context.Context) *verifyContext {
	if v := ctx.Value(verifyContextKey{}); v != nil {
		return v.(*verifyContext) //nolint:forcetypeassert
	}
	return nil
}
