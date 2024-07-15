package msgpack

import (
	"github.com/vmihailenco/msgpack/v5"

	"github.com/justenwalker/mack/encoding"
	"github.com/justenwalker/mack/macaroon"
)

var (
	_ encoding.MacaroonEncoder = EncoderDecoder{}
	_ encoding.MacaroonDecoder = EncoderDecoder{}
	_ encoding.StackEncoder    = EncoderDecoder{}
	_ encoding.StackDecoder    = EncoderDecoder{}
)

func (EncoderDecoder) DecodeStack(bs []byte, stack *macaroon.Stack) error {
	var raw stackEncoder
	if err := msgpack.Unmarshal(bs, &raw); err != nil {
		return err
	}
	*stack = raw.Stack
	return nil
}

func (EncoderDecoder) DecodeMacaroon(bs []byte, m *macaroon.Macaroon) error {
	var raw rawMacaroon
	if err := msgpack.Unmarshal(bs, &raw); err != nil {
		return err
	}
	*m = macaroon.NewFromRaw(raw.Raw)
	return nil
}

func (EncoderDecoder) EncodeMacaroon(m *macaroon.Macaroon) ([]byte, error) {
	return msgpack.Marshal(&macaroonEncoder{Macaroon: m})
}

func (EncoderDecoder) EncodeStack(stack macaroon.Stack) ([]byte, error) {
	return msgpack.Marshal(&stackEncoder{Stack: stack})
}
