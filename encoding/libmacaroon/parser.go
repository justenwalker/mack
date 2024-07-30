package libmacaroon

import (
	"errors"

	"github.com/justenwalker/mack/macaroon"
)

var errNoData = errors.New("libmacaroon: no macaroon data")

// Parser parses macaroons of the supported libmacaroon formats.
// it detects the encoding of format of the macaroon by inspecting the bytes.
// It expects v1 binary format to be base-64 encoded, as it is the canonical representation.
// All other formats should be in their canonical json or binary formats.
type Parser struct {
}

// DecodeMacaroon decodes a macaroon from the given binary or text data.
// The parser attempts to detect the format of the macaroon.
// The algorithm is as follows:
//
// 1. If the first byte is 0x02, then this is a binary formatted macaroon version 2.
//   - Parse as v2 and return
//
// 2. If the first byte is '{', then this is a json-formatted macaroon
//   - Try to interpret as v2j, and return it if it succeeds.
//   - Assume it is v1j and parse as that.
//
// 3. Assume binary format v1.
//
// NOTE: Since v1's "binary" format is canonically base-64 encoded, the parser expects v1
// macaroons to be encoded as base64.
func (v *Parser) DecodeMacaroon(bs []byte, m *macaroon.Macaroon) error {
	if len(bs) == 0 {
		return errNoData
	}
	if bs[0] == 2 { // version 2
		return (V2{}).DecodeMacaroon(bs, m)
	}
	if bs[0] == '{' { // json object
		if err := (V2J{}).DecodeMacaroon(bs, m); err == nil {
			return nil
		}
		return (V1J{}).DecodeMacaroon(bs, m)
	}
	data, err := Base64DecodeLoose(string(bs))
	if err != nil {
		return err
	}
	return (V1{}).DecodeMacaroon(data, m)
}

// DecodeStack decodes a stack macaroons from the given binary or text data.
// The parser attempts to detect the format of the macaroon.
// The algorithm is as follows:
//
// 1. If the first byte is 0x02, then this is a binary formatted macaroon version 2.
//   - Parse as v2 and return
//
// 2. If the first byte is '[', then this is a json-formatted list of macaroons
//   - Try to interpret as an array of v2j, and return it if it succeeds.
//   - Assume it is an array of v1j and parse as that.
//
// 3. Assume binary format v1.
//
// NOTE: Since v1's "binary" format is canonically base-64 encoded, the parser expects v1
// macaroons to be encoded as base64.
func (v *Parser) DecodeStack(bs []byte, stack *macaroon.Stack) error {
	if len(bs) == 0 {
		return errNoData
	}
	if bs[0] == 2 { // version 2
		return (V2{}).DecodeStack(bs, stack)
	}
	if bs[0] == '[' { // json array
		if err := (V2J{}).DecodeStack(bs, stack); err == nil {
			return nil
		}
		return (V1J{}).DecodeStack(bs, stack)
	}
	data, err := Base64DecodeLoose(string(bs))
	if err != nil {
		return err
	}
	return (V1{}).DecodeStack(data, stack)
}
