// Package thirdparty contains utilities for creating and discharging third party caveats.
//
// Third-Party caveats, unlike first-party, are cleared by a third-party service.
// A third party caveat is a tuple of 3 values: a Verification ID, the Caveat ID, and the Location hint
// (ie: where to obtain a discharge token).
//
// The process of adding a third party caveat to a macaroon is called Attenuation.
// When adding a third-party caveat, you require the bearer of the Authorizing macaroon to provide a corresponding
// Discharge Macaroon obtained from a third party.
//
// Client or servers attenuating macaroons may use the [Attenuator] to help construct the caveat.
// Servers may use the [Discharger] to extract a [Ticket] from the caveat id, which can be evaluated
// before generating a discharge macaroon.
//
// While the protocol for requesting a discharge token from a third party is not defined in the spec,
// if the implementation implements [ThirdParty], then they can be collected into a [Set], which allows
// can help locate all third-party caveats on a Macaroon and request a discharge macaroon from the corresponding [ThirdParty].
package thirdparty

import (
	"context"

	macaroon "github.com/justenwalker/mack"
)

// ThirdParty contacts a third-party service with a caveat ID and returns a Discharge Macaroon.
// How to request a discharge macaroon from a third party is not covered by the spec; but
// if the implementation satisfies this interface, then this package can be used to construct a [Set]
// to discharge all matching third party caveats on a macaroon.
type ThirdParty interface {
	// MatchCaveat returns true if this caveat can be discharged by this ThirdParty.
	MatchCaveat(c *macaroon.Caveat) bool
	// DischargeCaveat requests a discharge macaroon for the given caveat.
	DischargeCaveat(ctx context.Context, c *macaroon.Caveat) (macaroon.Macaroon, error)
}

// TicketExtractor extracts a Ticket from a third-party caveat id.
// This is the "dual" of the [CaveatIDIssuer]
//
// If the [CaveatIDIssuer] was implemented with the third party generating an opaque caveat id and associating it
// to a [Ticket] database, then this would look up that ticket and return it.
//
// If the [CaveatIDIssuer] instead encrypted the ticket for the third party using its public key, then
// this would extract the ticket from the caveat id be decrypting and decoding it.
type TicketExtractor interface {
	// ExtractTicket extracts a Ticket from an opaque third-party caveat id (cId).
	ExtractTicket(ctx context.Context, cID []byte) (*Ticket, error)
}

// PredicateChecker interprets a caveat id, and evaluates the result, and returns true if the predicate is satisfied.
type PredicateChecker interface {
	// CheckPredicate checks the given predicate, returning true if and only if the predicate holds.
	// The function MAY return an error, if it does, the error does not indicate the predicate is false,
	// but rather, that the predicate cannot be verified at this time.
	CheckPredicate(ctx context.Context, predicate []byte) (bool, error)
}

// CaveatIDIssuer issues caveat IDs to attenuate macaroons with third-party caveats.
//
// It exchanges a [Ticket] which is a pair containing a CaveatKey that is randomly generated for this
// macaroon, and a Predicate to be evaluated by the third party before discharges the caveat; for an opaque caveat ID
// that only the third party can use later to recover the Caveat Key and Predicate.
//
// One way to implement this is by having a third-party implement an API that can take this [Ticket] and
// return a caveat id. This requires the third party to be active in minting a caveat, and typically
// would create an cId/cK. Implementing such a protocol is out of scope for this library, but another library
// implementing [CaveatIDIssuer] may provide it.
//
// Another way is to use public-key cryptography to encrypt the caveat key/predicate payload.
// This doesn't require a third party to be an active participant in the creation of the caveat.
// Instead, the Caveat ID is constructed by encrypting the Caveat Key and Predicate using the third party's public key.
type CaveatIDIssuer interface {
	// IssueCaveatID issues an opaque caveat ID (cId) from a Ticket.
	IssueCaveatID(ctx context.Context, ticket Ticket) ([]byte, error)
}

// Set is a collection of third parties that can discharge third party caveats.
// While the protocol for requesting a discharge token from a third party is not defined in the spec,
// if the implementation implements [ThirdParty], then they can be collected into a [Set], which allows
// can help locate all third-party caveats on a Macaroon and request a discharge macaroon from the corresponding [ThirdParty].
type Set []ThirdParty

// Discharge iterates through all third party caveats and discharges them with the matching ThirdParty in the Set.
// If this discharge macaroon contains any third-party caveats, those too are discharged.
func (tps Set) Discharge(ctx context.Context, m *macaroon.Macaroon) ([]macaroon.Macaroon, error) {
	caveats := m.ThirdPartyCaveats()
	discharge := make([]macaroon.Macaroon, 0, len(caveats))
	for i := 0; i < len(caveats); i++ {
		if len(caveats[i].VID()) == 0 { // first party
			continue
		}
		dm, err := tps.dischargeCaveat(ctx, &caveats[i])
		if err != nil {
			return nil, err
		}
		discharge = append(discharge, dm)
		caveats = append(caveats, dm.ThirdPartyCaveats()...)
	}
	return discharge, nil
}

func (tps Set) dischargeCaveat(ctx context.Context, cp *macaroon.Caveat) (macaroon.Macaroon, error) {
	for _, tp := range tps {
		if !tp.MatchCaveat(cp) {
			continue
		}
		// matched
		dm, err := tp.DischargeCaveat(ctx, cp)
		if err != nil {
			return macaroon.Macaroon{}, &DischargeCaveatError{
				caveat: cp,
				err:    err,
			}
		}
		return dm, nil
	}
	return macaroon.Macaroon{}, &DischargeCaveatError{
		caveat: cp,
		err:    ErrNoMatchingThirdParty,
	}
}
