package encoding

import "github.com/justenwalker/mack"

// EncoderDecoder can encode and decode *mack.Macaroon and mack.Stack to and from byte representation.
type EncoderDecoder interface {
	MacaroonEncoder
	MacaroonDecoder
	StackEncoder
	StackDecoder
}

// MacaroonEncoder encodes a [mack.Stack] into bytes.
type MacaroonEncoder interface {
	EncodeMacaroon(m *mack.Macaroon) ([]byte, error)
}

// StackEncoder encodes a [mack.Stack] into bytes.
type StackEncoder interface {
	EncodeStack(stack mack.Stack) ([]byte, error)
}

// MacaroonDecoder decodes a [mack.Macaroon] from bytes.
type MacaroonDecoder interface {
	DecodeMacaroon(bs []byte, m *mack.Macaroon) error
}

// StackDecoder decodes a [mack.Macaroon] authorization stack from bytes.
type StackDecoder interface {
	DecodeStack(bs []byte, stack *mack.Stack) error
}
