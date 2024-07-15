// Package msgpack provides an Encoding for various macaroon types msgpack.
package msgpack

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/justenwalker/mack/exchange"
	"github.com/justenwalker/mack/macaroon"
	"github.com/justenwalker/mack/macaroon/thirdparty"
)

// Encoding implements the macaroon.Encoder, macaroon.Decoder, exchange.Encoder,
// and exchange.Decoder interfaces. It provides methods to encode and decode different types of structured
// data using msgpack.
var Encoding EncoderDecoder

// EncoderDecoder is a type that implements the macaroon.Encoder, macaroon.Decoder, exchange.Encoder,
// and exchange.Decoder interfaces. It provides methods to encode and decode different types of structured
// data using msgpack.
type EncoderDecoder struct{}

type rawMacaroon struct {
	macaroon.Raw
}

func (m *rawMacaroon) DecodeMsgpack(decoder *msgpack.Decoder) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("msgpack.DecodeMsgpack<Macaroon>: %w", err)
		}
	}()
	var mlen int
	if mlen, err = decoder.DecodeMapLen(); err != nil {
		return err
	}
	for i := 0; i < mlen; i++ {
		var key string
		key, err = decoder.DecodeString()
		if err != nil {
			return err
		}
		switch key {
		case "loc":
			var loc string
			loc, err = decoder.DecodeString()
			if err != nil {
				return err
			}
			m.Location = loc
		case "id":
			var id []byte
			id, err = decoder.DecodeBytes()
			if err != nil {
				return err
			}
			m.ID = id
		case "sig":
			var sig []byte
			sig, err = decoder.DecodeBytes()
			if err != nil {
				return err
			}
			m.Signature = sig
		case "caveats":
			var clen int
			clen, err = decoder.DecodeArrayLen()
			if err != nil {
				return err
			}
			m.Caveats = make([]macaroon.RawCaveat, clen)
			for c := 0; c < clen; c++ {
				var rc rawCaveat
				if err = rc.DecodeMsgpack(decoder); err != nil {
					return fmt.Errorf("[%d]: %w", c, err)
				}
				m.Caveats[c] = rc.RawCaveat
			}
		default:
			return fmt.Errorf("unknown map key '%s'", key)
		}
	}
	return nil
}

type rawCaveat struct {
	macaroon.RawCaveat
}

func (c *rawCaveat) DecodeMsgpack(decoder *msgpack.Decoder) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("msgpack.DecodeMsgpack<Caveat>: %w", err)
		}
	}()
	var mlen int
	if mlen, err = decoder.DecodeMapLen(); err != nil {
		return err
	}
	for i := 0; i < mlen; i++ {
		var key string
		key, err = decoder.DecodeString()
		if err != nil {
			return err
		}
		switch key {
		case "loc":
			var loc string
			loc, err = decoder.DecodeString()
			if err != nil {
				return err
			}
			c.Location = loc
		case "vid":
			var vid []byte
			vid, err = decoder.DecodeBytes()
			if err != nil {
				return err
			}
			c.VID = vid
		case "cid":
			var cid []byte
			cid, err = decoder.DecodeBytes()
			if err != nil {
				return err
			}
			c.CID = cid
		default:
			return fmt.Errorf("unknown map key '%s'", key)
		}
	}
	return nil
}

type macaroonEncoder struct {
	Macaroon *macaroon.Macaroon
}

func (m *macaroonEncoder) EncodeMsgpack(encoder *msgpack.Encoder) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("msgpack.EncodeMsgpack<Macaroon>: %w", err)
		}
	}()
	if err = encoder.EncodeMapLen(4); err != nil { // loc, id, caveats, sig
		return err
	}
	if err = encoder.EncodeString("loc"); err != nil {
		return err
	}
	if err = encoder.EncodeString(m.Macaroon.Location()); err != nil {
		return err
	}
	if err = encoder.EncodeString("id"); err != nil {
		return err
	}
	if err = encoder.EncodeBytes(m.Macaroon.ID()); err != nil {
		return err
	}
	if err = encoder.EncodeString("caveats"); err != nil {
		return err
	}
	cavs := m.Macaroon.Caveats()
	if err = encoder.EncodeArrayLen(len(cavs)); err != nil {
		return err
	}
	for _, c := range cavs {
		if err = encoder.Encode(&caveatEncoder{
			Caveat: c,
		}); err != nil {
			return err
		}
	}
	if err = encoder.EncodeString("sig"); err != nil {
		return err
	}
	if err = encoder.EncodeBytes(m.Macaroon.Signature()); err != nil {
		return err
	}
	return nil
}

type caveatEncoder struct {
	macaroon.Caveat
}

func (c *caveatEncoder) EncodeMsgpack(encoder *msgpack.Encoder) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("msgpack.EncodeMsgpack<Macaroon>: %w", err)
		}
	}()
	mlen := 1
	if len(c.VID()) > 0 {
		mlen = 3
	}
	if err = encoder.EncodeMapLen(mlen); err != nil { // loc, id, caveats, sig
		return err
	}
	if mlen == 3 {
		if err = encoder.EncodeString("loc"); err != nil {
			return err
		}
		if err = encoder.EncodeString(c.Location()); err != nil {
			return err
		}
		if err = encoder.EncodeString("vid"); err != nil {
			return err
		}
		if err = encoder.EncodeBytes(c.VID()); err != nil {
			return err
		}
	}
	if err = encoder.EncodeString("cid"); err != nil {
		return err
	}
	if err = encoder.EncodeBytes(c.ID()); err != nil {
		return err
	}
	return nil
}

type stackEncoder struct {
	Stack macaroon.Stack
}

func (c *stackEncoder) EncodeMsgpack(encoder *msgpack.Encoder) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("msgpack.EncodeMsgpack<Stack>: %w", err)
		}
	}()
	if err = encoder.EncodeArrayLen(len(c.Stack)); err != nil {
		return err
	}
	for i := range c.Stack {
		if err = encoder.Encode(&macaroonEncoder{Macaroon: &c.Stack[i]}); err != nil {
			return err
		}
	}
	return nil
}

func (c *stackEncoder) DecodeMsgpack(decoder *msgpack.Decoder) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("msgpack.DecodeMsgpack<Stack>: %w", err)
		}
	}()
	var alen int
	alen, err = decoder.DecodeArrayLen()
	if err != nil {
		return err
	}
	stack := make([]macaroon.Macaroon, alen)
	for i := 0; i < alen; i++ {
		var raw rawMacaroon
		if err = decoder.Decode(&raw); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
		stack[i] = macaroon.NewFromRaw(raw.Raw)
	}
	c.Stack = stack
	return nil
}

type encryptedMessageEncoder struct {
	EncryptedMessage exchange.EncryptedMessage
}

func (e *encryptedMessageEncoder) DecodeMsgpack(decoder *msgpack.Decoder) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("msgpack.DecodeMsgpack<EncryptedMessage>: %w", err)
		}
	}()
	var mlen int
	mlen, err = decoder.DecodeMapLen()
	if err != nil {
		return err
	}
	for i := 0; i < mlen; i++ {
		var key string
		key, err = decoder.DecodeString()
		if err != nil {
			return err
		}
		switch key {
		case "type":
			e.EncryptedMessage.Type, err = decoder.DecodeString()
			if err != nil {
				return err
			}
		case "kid":
			e.EncryptedMessage.KeyID, err = decoder.DecodeString()
			if err != nil {
				return err
			}
		case "data":
			e.EncryptedMessage.Payload, err = decoder.DecodeBytes()
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown map key '%s'", key)
		}
	}
	return nil
}

func (e *encryptedMessageEncoder) EncodeMsgpack(encoder *msgpack.Encoder) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("msgpack.EncodeMsgpack<EncryptedMessage>: %w", err)
		}
	}()
	if err = encoder.EncodeMapLen(3); err != nil {
		return err
	}
	if err = encoder.EncodeString("type"); err != nil {
		return err
	}
	if err = encoder.EncodeString(e.EncryptedMessage.Type); err != nil {
		return err
	}
	if err = encoder.EncodeString("kid"); err != nil {
		return err
	}
	if err = encoder.EncodeString(e.EncryptedMessage.KeyID); err != nil {
		return err
	}
	if err = encoder.EncodeString("data"); err != nil {
		return err
	}
	if err = encoder.EncodeBytes(e.EncryptedMessage.Payload); err != nil {
		return err
	}
	return nil
}

type ticketEncoder struct {
	thirdparty.Ticket
}

func (t *ticketEncoder) DecodeMsgpack(decoder *msgpack.Decoder) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("msgpack.DecodeMsgpack<Ticket>: %w", err)
		}
	}()
	var mlen int
	mlen, err = decoder.DecodeMapLen()
	for i := 0; i < mlen; i++ {
		var key string
		key, err = decoder.DecodeString()
		if err != nil {
			return err
		}
		switch key {
		case "ck":
			t.CaveatKey, err = decoder.DecodeBytes()
			if err != nil {
				return err
			}
		case "id":
			t.Predicate, err = decoder.DecodeBytes()
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown map key '%s'", key)
		}
	}
	return nil
}

func (t *ticketEncoder) EncodeMsgpack(encoder *msgpack.Encoder) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("msgpack.EncodeMsgpack<Ticket>: %w", err)
		}
	}()
	if err = encoder.EncodeMapLen(2); err != nil {
		return err
	}
	if err = encoder.EncodeString("ck"); err != nil {
		return err
	}
	if err = encoder.EncodeBytes(t.CaveatKey); err != nil {
		return err
	}
	if err = encoder.EncodeString("id"); err != nil {
		return err
	}
	if err = encoder.EncodeBytes(t.Predicate); err != nil {
		return err
	}
	return nil
}

var (
	_ msgpack.CustomEncoder = (*macaroonEncoder)(nil)
	_ msgpack.CustomEncoder = (*caveatEncoder)(nil)
	_ msgpack.CustomDecoder = (*rawMacaroon)(nil)
	_ msgpack.CustomDecoder = (*rawCaveat)(nil)
	_ msgpack.CustomEncoder = (*stackEncoder)(nil)
	_ msgpack.CustomDecoder = (*stackEncoder)(nil)
	_ msgpack.CustomEncoder = (*encryptedMessageEncoder)(nil)
	_ msgpack.CustomDecoder = (*encryptedMessageEncoder)(nil)
	_ msgpack.CustomEncoder = (*ticketEncoder)(nil)
	_ msgpack.CustomDecoder = (*ticketEncoder)(nil)
)
