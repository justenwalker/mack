package msgpack

import (
	"github.com/vmihailenco/msgpack/v5"

	"github.com/justenwalker/mack"
	"github.com/justenwalker/mack/encoding"
)

var _ encoding.EncoderDecoder = EncoderDecoder{}

// Encoding implements encoding.Encoding, exchange.Encoding interfaces.
// and exchange.Decoder interfaces. It provides methods to encode and decode different types of structured
// data using msgpack.
var Encoding EncoderDecoder //nolint:gochecknoglobals

// EncoderDecoder implements encoding.Encoding, exchange.Encoding interfaces.
// It provides methods to encode and decode different types of structured
// data using msgpack.
type EncoderDecoder struct{}

func (EncoderDecoder) DecodeStack(bs []byte, stack *mack.Stack) error {
	var raw stackEncoder
	if err := msgpack.Unmarshal(bs, &raw); err != nil {
		return err
	}
	*stack = raw.Stack
	return nil
}

func (EncoderDecoder) DecodeMacaroon(bs []byte, m *mack.Macaroon) error {
	var raw rawMacaroon
	if err := msgpack.Unmarshal(bs, &raw); err != nil {
		return err
	}
	*m = mack.NewFromRaw(raw.Raw)
	return nil
}

func (EncoderDecoder) EncodeMacaroon(m *mack.Macaroon) ([]byte, error) {
	return msgpack.Marshal(&macaroonEncoder{Macaroon: m})
}

func (EncoderDecoder) EncodeStack(stack mack.Stack) ([]byte, error) {
	return msgpack.Marshal(&stackEncoder{Stack: stack})
}
