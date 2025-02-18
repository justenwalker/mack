package mack

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strconv"
	"testing"
)

func BenchmarkNew(b *testing.B) {
	key := make([]byte, 32)
	id := make([]byte, 100)
	loc := make([]byte, 50)
	cavs := make([][]byte, 10)
	_, _ = rand.Read(key)
	_, _ = rand.Read(id)
	_, _ = rand.Read(loc)
	for i := range cavs {
		cavs[i] = make([]byte, 100)
		_, _ = rand.Read(cavs[i])
	}
	locStr := string(loc)
	sch, _ := NewScheme(SchemeConfig{
		HMACScheme:           testScheme{},
		EncryptionScheme:     testScheme{},
		BindForRequestScheme: testScheme{},
	})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err := sch.NewMacaroon(locStr, id, key, cavs...)
		if err != nil {
			b.Fatalf("newMacaroon: %v", err)
		}
	}
}

func BenchmarkMacaroon_Clone(b *testing.B) {
	caveats := []int{1, 10, 100}
	for _, n := range caveats {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			m := helpGenerateMacaroon(b, n)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				m = Clone(&m)
			}
		})
	}
}

func BenchmarkMacaroon_AddCaveat(b *testing.B) {
	caveats := []int{1, 10, 100}
	sch, err := NewScheme(SchemeConfig{
		HMACScheme:           testScheme{},
		EncryptionScheme:     testScheme{},
		BindForRequestScheme: testScheme{},
	})
	if err != nil {
		b.Fatalf("failed to create new scheme: %v", err)
	}
	for _, n := range caveats {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			m := helpGenerateMacaroon(b, n)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				m, err = sch.AddFirstPartyCaveat(&m, []byte(`9d864f2248e7401eaf01e07032bb18469d864f2248e7401eaf01e07032bb1846`))
			}
		})
	}
}

func helpGenerateMacaroon(tb testing.TB, caveats int) Macaroon {
	tb.Helper()
	raw := Raw{
		ID:        []byte(`9d864f22-48e7-401e-af01-e07032bb1846`),
		Location:  "https://example.org",
		Caveats:   make([]RawCaveat, caveats),
		Signature: []byte(`9d864f2248e7401eaf01e07032bb18469d864f2248e7401eaf01e07032bb1846`),
	}
	for i := range raw.Caveats {
		raw.Caveats[i] = RawCaveat{
			CID: []byte(`9d864f22-48e7-401e-af01-e07032bb1846`),
		}
	}
	return NewFromRaw(raw)
}

func TestMacaroon_Clone_allocs(t *testing.T) {
	m := helpGenerateMacaroon(t, 100)
	var result Macaroon
	allocs := testing.AllocsPerRun(10*1024, func() {
		result = Clone(&m)
	})
	const expected = 1 // copy buffer
	if allocs > expected {
		writeHeapProfile(t)
		t.Fatalf("allocs: %d > %d", int(allocs), expected)
	}
	t.Logf("AllocsPerRun: %d", int(allocs))
	_, _ = io.Discard.Write([]byte(fmt.Sprintf("%v", result)))
}

func TestMacaroon_AddCaveat_allocs(t *testing.T) {
	m := helpGenerateMacaroon(t, 100)
	sch, err := NewScheme(SchemeConfig{
		HMACScheme:           testScheme{},
		EncryptionScheme:     testScheme{},
		BindForRequestScheme: testScheme{},
	})
	if err != nil {
		t.Fatalf("failed to create new scheme: %v", err)
	}
	allocs := testing.AllocsPerRun(10*1024, func() {
		_, err = sch.AddFirstPartyCaveat(&m, []byte(`9d864f2248e7401eaf01e07032bb18469d864f2248e7401eaf01e07032bb1846`))
		if err != nil {
			t.Fatalf("unexpected error: %v/%v", err, errors.Unwrap(err))
		}
	})
	const expected = 1 // copy buffer
	if allocs > expected {
		writeHeapProfile(t)
		t.Fatalf("allocs: %d > %d", int(allocs), expected)
	}
	t.Logf("AllocsPerRun: %d", int(allocs))
}

type testScheme struct{}

func (t testScheme) BindForRequest(_ *Macaroon, _ []byte) error {
	return nil
}

func (t testScheme) Overhead() int {
	return 0
}

func (t testScheme) NonceSize() int {
	return 0
}

func (t testScheme) Encrypt(_ []byte, in []byte, _ []byte) ([]byte, error) {
	return in, nil
}

func (t testScheme) Decrypt(_ []byte, in []byte, _ []byte) ([]byte, error) {
	return in, nil
}

func (t testScheme) KeySize() int {
	return 32
}

func (t testScheme) HMAC(key []byte, out []byte, _ []byte) error {
	copy(key, out) // simulate something
	return nil
}

var (
	_ HMACScheme           = testScheme{}
	_ EncryptionScheme     = testScheme{}
	_ BindForRequestScheme = testScheme{}
)

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
