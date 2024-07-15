package encoding

import "github.com/justenwalker/mack/macaroon"

// MacaroonEncoder encodes a macaroon into bytes.
type MacaroonEncoder interface {
	EncodeMacaroon(m *macaroon.Macaroon) ([]byte, error)
}

// StackEncoder encodes a macaroon stack into bytes.
type StackEncoder interface {
	EncodeStack(stack macaroon.Stack) ([]byte, error)
}

// MacaroonDecoder decodes a macaroon from bytes.
type MacaroonDecoder interface {
	DecodeMacaroon(bs []byte, m *macaroon.Macaroon) error
}

// StackDecoder decodes a macaroon authorization stack from bytes.
type StackDecoder interface {
	DecodeStack(bs []byte, stack *macaroon.Stack) error
}
