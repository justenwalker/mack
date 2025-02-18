package mack

import (
	"errors"
	"fmt"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrPredicateNotSatisfied = Error("macaroon: predicate not satisfied")
	ErrVerificationFailed    = Error("macaroon: verification failed")
	ErrInvalidArgument       = Error("macaroon: invalid argument")
)

type predicateNotSatisfiedError struct {
	predicate Predicate
}

func (m *predicateNotSatisfiedError) Error() string {
	return fmt.Sprintf("%s: %v", ErrPredicateNotSatisfied, m.predicate)
}

func (m *predicateNotSatisfiedError) Predicate() Predicate {
	return m.predicate
}

func (m *predicateNotSatisfiedError) Is(err error) bool {
	return errors.Is(err, ErrPredicateNotSatisfied)
}

type verificationError struct {
	macaroon *Macaroon
	err      error
}

func (m *verificationError) Error() string {
	return string(ErrVerificationFailed)
}

func (m *verificationError) Macaroon() *Macaroon {
	return m.macaroon
}

func (m *verificationError) Is(err error) bool {
	if errors.Is(err, ErrVerificationFailed) {
		return true
	}
	return errors.Is(m.err, err)
}

func (m *verificationError) Unwrap() error {
	return m.err
}

func validationError(m *Macaroon, err error) error {
	return &verificationError{
		macaroon: m,
		err:      err,
	}
}
