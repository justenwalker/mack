package macaroon

// Raw represents raw macaroon data, used to construct a Macaroon from bytes with [NewFromRaw].
type Raw struct {
	ID        []byte
	Location  string
	Caveats   []RawCaveat
	Signature []byte
}

// RawCaveat represents raw caveat data.
type RawCaveat struct {
	CID      []byte
	VID      []byte
	Location string
}

func (c *RawCaveat) size() int {
	return len(c.CID) + len(c.VID) + len(c.Location)
}

// NewFromRaw creates a new Macaroon with the given Raw macaroon data.
// The raw macaroon data consists of the ID, location, caveats, and signature.
// The function copies the raw macaroon data into a new Macaroon instance.
// There is no validation done on the resulting Macaroon; this is only useful for Decoding a macaroon from a wire format.
func NewFromRaw(raw Raw) Macaroon {
	nmd := newMacaroonData(raw.Location, raw.ID, len(raw.Signature), raw.Caveats...)
	copy(nmd.sig(), raw.Signature)
	return Macaroon{data: nmd}
}
