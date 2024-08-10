package macaroon

import (
	"bytes"
	"crypto/hmac"
	"encoding/hex"
	"fmt"
)

type Macaroon struct {
	caveats []Caveat

	// data contains the full content of the macaroon and its caveats
	data *macaroonData
}

// newMacaroon creates a new Macaroon with the given scheme, key, id, and location.
func newMacaroon(s *Scheme, key []byte, id []byte, loc string) (Macaroon, error) {
	// newMacaroon(k, id , L)
	//	sig := MAC(k, id )
	//	return macaroon@L〈id , [ ], sig〉
	var m Macaroon
	m.data = newMacaroonData(loc, id, s.keySize)
	if err := s.hmac.HMAC(key, m.data.sig(), id); err != nil {
		return Macaroon{}, err
	}
	return m, nil
}

// Clone create a clone of the given Macaroon.
// If the macaroon is nil, then the Zero value is returned.
func Clone(m *Macaroon) Macaroon {
	if m == nil {
		return Macaroon{}
	}
	return Macaroon{data: m.data.clone()}
}

// IsZero returns true if the macaroon represents the Zero-value macaroon.
func (m *Macaroon) IsZero() bool {
	if m == nil {
		return true
	}
	return m.data == nil
}

// Location returns the string representation of the Location of the Macaroon.
func (m *Macaroon) Location() string {
	return m.data.loc()
}

// ID returns the ID of the macaroon.
func (m *Macaroon) ID() []byte {
	return m.data.id()
}

// Signature the signature of the Macaroon.
func (m *Macaroon) Signature() []byte {
	return m.data.sig()
}

// Caveats returns all macaroon caveats.
func (m *Macaroon) Caveats() []Caveat {
	if m.caveats == nil {
		m.caveats = m.data.caveats()
	}
	return m.caveats
}

func (m *Macaroon) caveatAt(i int) *Caveat {
	cs := m.Caveats()
	if len(cs) == 0 {
		return nil
	}
	return &cs[i]
}

// Equal tests if a macaroon is exactly equal to another macaroon.
func (m *Macaroon) Equal(o *Macaroon) bool {
	if m == o { // both nil, or identical pointers
		return true
	}
	if o == nil || m == nil {
		return false
	}
	return m.data.equal(o.data)
}

// FirstPartyCaveats returns a slice of first-party caveats from the Macaroon.
// A first-party caveat is a caveat that does not include a verifier ID.
// First-party caveats are evaluated by the target service.
func (m *Macaroon) FirstPartyCaveats() []Caveat {
	fpc := make([]Caveat, 0, len(m.Caveats()))
	for _, c := range m.Caveats() {
		if !c.thirdParty() {
			fpc = append(fpc, c)
		}
	}
	return fpc
}

// ThirdPartyCaveats returns an array of caveats that are third-party caveats.
// A third-party caveat is a caveat that includes a verifier ID.
// Third-party caveats are discharged by a third-party service with a Discharge Macaroon.
func (m *Macaroon) ThirdPartyCaveats() []Caveat {
	fpc := make([]Caveat, 0, len(m.Caveats()))
	for _, c := range m.Caveats() {
		if c.thirdParty() {
			fpc = append(fpc, c)
		}
	}
	return fpc
}

func (m *Macaroon) addFirstPartyCaveat(s *Scheme, predicate []byte) Macaroon {
	return m.addCaveats(s, RawCaveat{
		CID: predicate,
	})
}

func (m *Macaroon) addThirdPartyCaveat(s *Scheme, cKey []byte, cID []byte, cLoc string) (Macaroon, error) {
	vID, err := s.encrypt(nil, cKey, m.data.sig())
	if err != nil {
		return Macaroon{}, err
	}
	return m.addCaveats(s, RawCaveat{
		CID:      cID,
		VID:      vID,
		Location: cLoc,
	}), nil
}

func (m *Macaroon) addCaveats(s *Scheme, rcs ...RawCaveat) Macaroon {
	nmd := m.data.appendCaveats(s, rcs...)
	return Macaroon{
		data: nmd,
	}
}

// verify the macaroon.
// The key is the key secret key used to create the root macaroon.
// the sig is an optional buffer used for calculating the signatures, if not provided, it will allocate a buffer.
func (m *Macaroon) verify(s *Scheme, stack Stack, key []byte, sigbuf []byte, v *verifyContext, vi int, discharged []byte) error {
	var err error
	defer func() {
		if err != nil {
			v.fail(vi, err)
		}
	}()
	if len(key) != s.keySize {
		return fmt.Errorf("%w: invalid key size. need=%d, got=%d", ErrInvalidArgument, s.keySize, len(key))
	}
	if len(sigbuf) != s.keySize {
		sigbuf = make([]byte, s.keySize)
	}
	vo := v.traceRootKey(vi, key, m.ID())
	if err = s.hmac.HMAC(key, sigbuf, m.ID()); err != nil {
		return fmt.Errorf("error executing hmac: %w", err)
	}
	vo.setResult(sigbuf)
	for i := range m.Caveats() {
		err = m.verifyCaveat(s, stack, sigbuf, m.caveatAt(i), v, vi, discharged)
		if err != nil {
			return err
		}
	}
	target := stack.Target()
	if m != target {
		vo = v.trace(vi, TraceOpBind, target.Signature(), sigbuf)
		err = s.bfr.BindForRequest(target, sigbuf)
		vo.setResult(sigbuf)
		if err != nil {
			return validationError(m, fmt.Errorf("macaroon.verify: could not get request signature: %w", err))
		}
	}
	if !hmac.Equal(m.data.sig(), sigbuf) {
		return validationError(m, fmt.Errorf("macaroon.verify: signatures did not match: want=%s, got=%s", hex.EncodeToString(sigbuf), hex.EncodeToString(m.data.sig())))
	}
	return nil
}

func (m *Macaroon) verifyCaveat(s *Scheme, stack Stack, cSig []byte, c *Caveat, v *verifyContext, vi int, discharged []byte) error {
	if len(c.VID()) == 0 { // first party
		vo := v.trace(vi, TraceOpHMAC, cSig, c.data())
		err := s.hmac.HMAC(cSig, cSig, c.data())
		vo.setResult(cSig)
		return err
	}
	cK := s.getKeyBuffer()
	defer s.releaseKeyBuffer(cK)
	vo := v.trace(vi, TraceOpDecrypt, cSig, c.VID())
	_, err := s.decrypt(*cK, c.VID(), cSig)
	vo.setResult(*cK)
	if err != nil {
		return validationError(m, fmt.Errorf("macaroon.Caveat: failed to decrypt third-party caveat verification key: %w", err))
	}
	discharges := stack.Discharges()
	for i := range discharges {
		if bytes.Equal(discharges[i].ID(), c.ID()) {
			if discharged[i] < 255 {
				discharged[i]++
			}
			if err = discharges[i].verify(s, stack, *cK, *cK, v, i+1, discharged); err != nil {
				return err
			}
			vo := v.trace(vi, TraceOpHMAC, cSig, c.data())
			err = s.hmac.HMAC(cSig, cSig, c.data())
			vo.setResult(cSig)
			return err
		}
	}
	return validationError(m, fmt.Errorf("macaroon.Caveat: missing discharge for caveat: %v", c.ID()))
}

func (m *Macaroon) bindForRequest(s *Scheme, tm *Macaroon) error {
	return s.bfr.BindForRequest(tm, m.data.sig())
}
