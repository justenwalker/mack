package sensible

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"testing"

	"github.com/justenwalker/mack"
	"github.com/justenwalker/mack/internal/testhelpers"
)

func TestSensibleScheme_Validate(t *testing.T) {
	ctx := context.Background()
	ts, stack := helperCreate3PMacaroon(t)
	_, err := ts.Verify(ctx, testhelpers.RootKey, stack)
	if err != nil {
		t.Logf("m: %v", stack.Target())
		for _, d := range stack.Discharges() {
			t.Logf("d: %v", &d)
		}
		t.Fatalf("Clear: %v", errors.Unwrap(err))
	}
}

func TestSensibleScheme_Validate_allocs(t *testing.T) {
	ctx := context.Background()
	ts, stack := helperCreate3PMacaroon(t)
	allocs := testing.AllocsPerRun(10*1024, func() {
		_, err := ts.Verify(ctx, testhelpers.RootKey, stack)
		if err != nil {
			t.Fatalf("unexpected error: %v/%v", err, errors.Unwrap(err))
		}
	})
	const expected = 2
	t.Logf("AllocsPerRun: %d", int(allocs))
	if allocs > expected {
		writeHeapProfile(t)
		t.Fatalf("allocs: %d > %d", int(allocs), expected)
	}
	t.Logf("AllocsPerRun: %d", int(allocs))
}

func writeHeapProfile(t *testing.T) {
	t.Helper()
	var err error
	filename := filepath.Join("testdata", t.Name()+".heap.out")
	var fd *os.File
	fd, err = os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0o644))
	if err != nil {
		t.Errorf("writeHeapProfile: failed to craete memory profile file '%s': %v", filename, err)
		return
	}
	defer func() {
		if err == nil {
			return
		}
		// make sure its closed
		_ = fd.Close()
		// try to remove
		if rmErr := os.Remove(filename); rmErr != nil {
			t.Errorf("writeHeapProfile: failed to remove file '%s': %v", filename, err)
			return
		}
	}()
	if err = pprof.Lookup("heap").WriteTo(fd, 0); err != nil {
		t.Errorf("writeHeapProfile: failed to write heap profile '%s': %v", filename, err)
		return
	}
	if err = fd.Close(); err != nil {
		t.Errorf("writeHeapProfile: failed to close heap profile '%s': %v", filename, err)
		return
	}
	t.Logf("memory profile written to: %v", filename)
}

var errBenchmarkSchemeVerify error

func BenchmarkSensibleScheme_Validate(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()
	ts, stack := helperCreate3PMacaroon(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, errBenchmarkSchemeVerify = ts.Verify(ctx, testhelpers.RootKey, stack)
	}
	_, _ = fmt.Fprintln(io.Discard, errBenchmarkSchemeVerify)
}

func helperCreate3PMacaroon(tb testing.TB) (*mack.Scheme, mack.Stack) {
	tb.Helper()
	sch := Scheme()
	m, err := sch.UnsafeRootMacaroon("1p", []byte("hello"), testhelpers.RootKey)
	if err != nil {
		tb.Fatalf("UnsafeRootMacaroon: %v", err)
	}
	m, _ = sch.AddFirstPartyCaveat(&m, []byte(`a > 1`))
	m, _ = sch.AddFirstPartyCaveat(&m, []byte(`b > 2`))
	m, err = sch.AddThirdPartyCaveat(&m, testhelpers.ThirdPartyKey, []byte("{cK,userid == foo}"), "https://other.example.org")
	if err != nil {
		tb.Fatalf("AddThirdPartyCaveat: %v", err)
	}
	m, _ = sch.AddFirstPartyCaveat(&m, []byte(`user = foo`))
	dm, err := sch.UnsafeRootMacaroon("https://other.example.org", []byte("{cK,userid == foo}"), testhelpers.ThirdPartyKey)
	if err != nil {
		tb.Fatalf("UnsafeRootMacaroon: %v", err)
	}
	stack, err := sch.PrepareStack(&m, []mack.Macaroon{dm})
	if err != nil {
		tb.Fatalf("BindForRequest: %v", err)
	}
	return sch, stack
}
