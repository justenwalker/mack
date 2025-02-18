package mack

import (
	"context"
	"fmt"
)

// Stack is a slice of [Macaroon] which represent the authorizing macaroon (Target) and all discharge macaroons bound to it.
// The first element in the slice is the authorizing macaroon. The remaining macaroons are discharge macaroons which
// have had their signatures finalized by binding them to the authorizing macaroon.
type Stack []Macaroon

// Target returns the authorizing macaroon.
func (s Stack) Target() *Macaroon {
	return &s[0]
}

// Discharges returns all the discharge macaroons.
func (s Stack) Discharges() []Macaroon {
	return s[1:]
}

// InsecureVerifiedStack converts the [Stack] into a [VerifiedStack] without actually verifying it.
// It may be useful if the verification was already done on the [Stack] before and that result was cached.
// Generally, it is insecure to use this, instead use the [Scheme.Verify] function.
func InsecureVerifiedStack(stack Stack) *VerifiedStack {
	return &VerifiedStack{
		stack: stack,
	}
}

// VerifiedStack represents a stack of macaroons which have had their cryptographic signatures verified.
type VerifiedStack struct {
	verified bool
	stack    Stack
}

// ID returns the authorization macaroon ID.
func (v *VerifiedStack) ID() []byte {
	return v.stack[0].ID()
}

// Verified returns true if the stack was verified.
// A InsecureVerifiedStack will return false.
func (v *VerifiedStack) Verified() bool {
	return v.verified
}

// Predicates returns the list of first-party predicates found in the verified stack.
func (v *VerifiedStack) Predicates() []Predicate {
	var predicates []Predicate
	for i := range v.stack {
		predicates = append(predicates, caveatsToPredicates(&v.stack[i])...)
	}
	return predicates
}

func caveatsToPredicates(m *Macaroon) []Predicate {
	cs := m.FirstPartyCaveats()
	predicates := make([]Predicate, len(cs))
	for i := range cs {
		predicates[i] = Predicate{
			MacaroonID: m.ID(),
			CaveatID:   cs[i].ID(),
			Index:      i,
		}
	}
	return predicates
}

// Clear clears the verified stack by checking each macaroon in the stack. It uses the provided PredicateChecker
// to verify each predicate in the stack of macaroons. If a predicate is not satisfied, this will return an error.
// If a predicate was not satisfied, this condition may be detected with errors.Is(err, ErrPredicateNotSatisfied)
// The resulting error should also have a Predicate() function, that returns the predicate which failed.
func (v *VerifiedStack) Clear(ctx context.Context, pcheck PredicateChecker) error {
	for i := range v.stack {
		if err := checkMacaroon(ctx, &v.stack[i], pcheck); err != nil {
			return err
		}
	}
	return nil
}

func checkMacaroon(ctx context.Context, m *Macaroon, pcheck PredicateChecker) error {
	for i := range m.Caveats() {
		if m.caveatAt(i).thirdParty() {
			continue
		}
		predicate := Predicate{
			MacaroonID: m.ID(),
			CaveatID:   m.caveatAt(i).ID(),
			Index:      i,
		}
		ok, err := pcheck.CheckPredicate(ctx, predicate.CaveatID)
		if err != nil {
			return fmt.Errorf("macaroon.Caveat: failed to verify caveat '%v': %w", &predicate, err)
		}
		if !ok {
			return &predicateNotSatisfiedError{
				predicate: predicate,
			}
		}
	}
	return nil
}

// Predicate is a struct that contains the Macaroon ID, Caveat ID, and the position in the macaroon which it was fond.
type Predicate struct {
	MacaroonID []byte
	CaveatID   []byte
	Index      int
}

func (p Predicate) String() string {
	return fmt.Sprintf("/macaroon/%s/caveat/%d: %s", printableBytes(p.MacaroonID), p.Index, printableBytes(p.CaveatID))
}

// PredicateChecker verifies a predicate.
type PredicateChecker interface {
	// CheckPredicate checks the given predicate, returning true if and only if the predicate holds.
	// The function MAY return an error, if it does, the error does not indicate the predicate is false,
	// but rather, that the predicate cannot be verified at this time.
	CheckPredicate(ctx context.Context, predicate []byte) (bool, error)
}
