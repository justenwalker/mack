package macaroon

import (
	"bytes"
	"testing"
)

func TestNewFromRaw(t *testing.T) {
	raw := Raw{
		ID:       []byte(`id`),
		Location: "location",
		Caveats: []RawCaveat{
			{
				CID: []byte("cav1"),
			},
			{
				CID: []byte("cav2"),
			},
			{
				CID:      []byte("cav3p"),
				VID:      []byte(`vid3p`),
				Location: "3p",
			},
		},
		Signature: []byte(`sig123`),
	}
	m := NewFromRaw(raw)
	if !bytes.Equal(raw.ID, m.ID()) {
		t.Errorf("id mismatch: got %v want %v", m.ID(), raw.ID)
	}
	if raw.Location != m.Location() {
		t.Errorf("location mismatch: got %v want %v", m.Location(), raw.Location)
	}
	if !bytes.Equal(raw.Signature, m.Signature()) {
		t.Errorf("signature mismatch: got %v want %v", m.Signature(), raw.Signature)
	}
	cavs := m.Caveats()
	if len(raw.Caveats) != len(cavs) {
		t.Fatalf("number of caveats mismatch, got %d want %d", len(raw.Caveats), len(cavs))
	}
	for i, a := range raw.Caveats {
		b := cavs[i]
		if !bytes.Equal(a.CID, b.ID()) {
			t.Errorf("caveat[%d]: cID mismatch: got %v want %v", i, b.ID(), a.CID)
		}
		if !bytes.Equal(a.VID, b.VID()) {
			t.Errorf("caveat[%d]: vID mismatch: got %v want %v", i, b.VID(), a.VID)
		}
		if a.Location != b.Location() {
			t.Errorf("caveat[%d]: cLoc mismatch: got %v want %v", i, b.Location(), a.Location)
		}
	}
}
