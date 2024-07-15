package thirdparty

import (
	"context"
	"errors"
	"fmt"

	"github.com/justenwalker/mack/crypt/random"
	"github.com/justenwalker/mack/macaroon"
)

// AttenuatorConfig configures a Attenuator.
type AttenuatorConfig struct {
	// Location is the location that is tagged on third-party caveats generated for this service.
	Location string
	// Scheme is the cryptographic scheme used for Macaroons (Required).
	Scheme *macaroon.Scheme
	// CaveatIssuer is used to issue caveat ids (Required).
	CaveatIssuer CaveatIDIssuer
}

// Attenuator applies third party caveats to a Macaroon.
// How caveat IDs are not covered by the spec, but if the implementation satisfies [CaveatIDIssuer]
// then an [Attenuator] may be used to apply caveats to a Macaroon.
type Attenuator struct {
	issuer   CaveatIDIssuer
	scheme   *macaroon.Scheme
	location string
}

// Ticket contains the third-party caveat root key and associated predicate.
// This is used when constructing a third-party caveat id (cId) to create attenuate a macaroon
// and extracted by the discharging service to be converted into a cId.
type Ticket struct {
	CaveatKey []byte
	Predicate []byte
}

// NewAttenuator creates a new instance of Attenuator.
func NewAttenuator(cfg AttenuatorConfig) (*Attenuator, error) {
	if cfg.Scheme == nil {
		return nil, errors.New("cfg.Scheme is nil")
	}
	if cfg.CaveatIssuer == nil {
		return nil, errors.New("cfg.CaveatIDIssuer is nil")
	}
	return &Attenuator{
		issuer:   cfg.CaveatIssuer,
		scheme:   cfg.Scheme,
		location: cfg.Location,
	}, nil
}

// Location returns the location of the third-party service.
func (a *Attenuator) Location() string {
	return a.location
}

// Attenuate adds a third-party caveat to a macaroon.
// It generates a new random key appends a third-party caveat with a `cId` issued by the CaveatIDIssuer.
func (a *Attenuator) Attenuate(ctx context.Context, m *macaroon.Macaroon, predicate []byte) (am macaroon.Macaroon, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("thirdparty.Attenuate: %w", err)
		}
	}()
	cKey := make([]byte, a.scheme.KeySize())
	random.Read(cKey)

	cID, err := a.issuer.IssueCaveatID(ctx, Ticket{
		CaveatKey: cKey,
		Predicate: predicate,
	})
	if err != nil {
		return am, err
	}
	am, err = a.scheme.AddThirdPartyCaveat(m, cKey, cID, a.location)
	if err != nil {
		return am, err
	}
	return am, nil
}
