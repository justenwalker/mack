package thirdparty

import (
	"context"
	"errors"

	"github.com/justenwalker/mack/macaroon"
)

// DischargerConfig contains the configuration options for a Discharger.
type DischargerConfig struct {
	// Location is the location that is used for the discharge macaroons.
	Location string
	// Scheme is the cryptographic scheme used for Macaroons (Required).
	Scheme *macaroon.Scheme
	// TicketExtractor extracts a ticket from a caveat id
	TicketExtractor TicketExtractor
}

// NewDischarger creates a new Discharger with the specified configuration.
//
// - If cfg.Scheme and cfg.Decryptor are mandatory. If not provided, an error is returned.
// - If cfg.Decoder is nil, the DefaultDecoder will be used.
func NewDischarger(cfg DischargerConfig) (*Discharger, error) {
	if cfg.Location == "" {
		return nil, errors.New("cfg.Location is empty")
	}
	if cfg.Scheme == nil {
		return nil, errors.New("cfg.Scheme is nil")
	}
	if cfg.TicketExtractor == nil {
		return nil, errors.New("cfg.TicketExtractor is nil")
	}
	return &Discharger{
		scheme:    cfg.Scheme,
		extractor: cfg.TicketExtractor,
		location:  cfg.Location,
	}, nil
}

// Discharger generates discharge tokens by decoding the caveat id and validating the predicate contained therein.
type Discharger struct {
	scheme    *macaroon.Scheme
	extractor TicketExtractor
	location  string
}

// Discharge generates a discharge token by extracting the key from the caveat ID,
// verifying the predicate, and creating a new macaroon.
func (d *Discharger) Discharge(ctx context.Context, cID []byte, pcheck PredicateChecker) (m macaroon.Macaroon, err error) {
	t, err := d.extractor.ExtractTicket(ctx, cID)
	if err != nil {
		return m, err
	}
	ok, err := pcheck.CheckPredicate(ctx, t.Predicate)
	if err != nil {
		return m, err
	}
	if !ok {
		return m, macaroon.ErrPredicateNotSatisfied
	}
	// This is safe because this macaroon is used only to discharge another macaroon with caveats
	return d.scheme.UnsafeRootMacaroon(d.location, cID, t.CaveatKey)
}
