package macaroon_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/justenwalker/mack/internal/testhelpers"
)

func TestCreateMacaroon(t *testing.T) {
	s := testhelpers.NewScheme(t)
	key := make([]byte, s.KeySize())
	id := []byte(`id`)
	loc := "loc"
	m, err := s.UnsafeRootMacaroon(loc, id, key)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(id, m.ID()); diff != "" {
		t.Fatalf("id mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(loc, m.Location()); diff != "" {
		t.Fatalf("location mismatch (-want +got):\n%s", diff)
	}
	cID := []byte("a")
	m, _ = s.AddFirstPartyCaveat(&m, cID)
	cs := m.Caveats()
	if len(cs) != 1 {
		t.Fatalf("expected to have 1 caveat after adding, instead got %d", len(cs))
	}
	if diff := cmp.Diff(cID, cs[0].ID()); diff != "" {
		t.Fatalf("caveat[0].ID mismatch (-want +got):\n%s", diff)
	}
	cID2 := []byte("b")
	cLoc2 := "3p"
	cKey := make([]byte, s.KeySize())
	m, err = s.AddThirdPartyCaveat(&m, cKey, cID2, cLoc2)
	if err != nil {
		t.Fatal(err)
	}
	cs = m.Caveats()
	if len(cs) != 2 {
		t.Fatalf("expected to have 2 caveats after adding, instead got %d", len(cs))
	}
	if diff := cmp.Diff(cID2, cs[1].ID()); diff != "" {
		t.Fatalf("caveat[1].ID mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(cLoc2, cs[1].Location()); diff != "" {
		t.Fatalf("caveat[1].Location mismatch (-want +got):\n%s", diff)
	}
	t.Log(m.String())
}
