package msgpack

import (
	"github.com/vmihailenco/msgpack/v5"

	"github.com/justenwalker/mack/thirdparty"
	"github.com/justenwalker/mack/thirdparty/exchange"
)

var _ exchange.EncoderDecoder = EncoderDecoder{}

func (EncoderDecoder) EncodeMessage(em *exchange.EncryptedMessage) ([]byte, error) {
	return msgpack.Marshal(&encryptedMessageEncoder{EncryptedMessage: *em})
}

func (EncoderDecoder) EncodeTicket(t thirdparty.Ticket) ([]byte, error) {
	return msgpack.Marshal(&ticketEncoder{Ticket: t})
}

func (EncoderDecoder) DecodeMessage(msg []byte) (*exchange.EncryptedMessage, error) {
	var enc encryptedMessageEncoder
	if err := msgpack.Unmarshal(msg, &enc); err != nil {
		return nil, err
	}
	return &enc.EncryptedMessage, nil
}

func (EncoderDecoder) DecodeTicket(bs []byte) (*thirdparty.Ticket, error) {
	var enc ticketEncoder
	if err := msgpack.Unmarshal(bs, &enc); err != nil {
		return nil, err
	}
	return &enc.Ticket, nil
}
