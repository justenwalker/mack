package thirdparty

import (
	macaroon "github.com/justenwalker/mack"
)

type stringError string

func (e stringError) Error() string {
	return string(e)
}

const ErrNoMatchingThirdParty = stringError("no matching third party for caveat")

// DischargeCaveatError is returned when there is an error discharging a caveat from a ThirdParty.
type DischargeCaveatError struct {
	caveat *macaroon.Caveat
	err    error
}

func (d *DischargeCaveatError) Caveat() macaroon.Caveat {
	return *d.caveat
}

func (d *DischargeCaveatError) Error() string {
	return "thirdparty: discharging failed"
}

func (d *DischargeCaveatError) Unwrap() error {
	return d.err
}
