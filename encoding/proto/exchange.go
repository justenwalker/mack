package proto

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/justenwalker/mack/encoding/proto/pb"
	"github.com/justenwalker/mack/exchange"
	"github.com/justenwalker/mack/macaroon/thirdparty"
)

var (
	_ exchange.Encoder = EncoderDecoder{}
	_ exchange.Decoder = EncoderDecoder{}
)

func (EncoderDecoder) EncodeMessage(em *exchange.EncryptedMessage) ([]byte, error) {
	bs, err := proto.Marshal(&pb.EncryptedMessage{
		Type:    em.Type,
		KeyId:   em.KeyID,
		Payload: em.Payload,
	})
	if err != nil {
		return nil, fmt.Errorf("protobufencode.EncodeMessage: %w", err)
	}
	return bs, nil
}

func (EncoderDecoder) EncodeTicket(t thirdparty.Ticket) ([]byte, error) {
	bs, err := proto.Marshal(&pb.Ticket{
		Key:       t.CaveatKey,
		Predicate: t.Predicate,
	})
	if err != nil {
		return nil, fmt.Errorf("protobufencode.EncodeTicket: %w", err)
	}
	return bs, nil
}

func (EncoderDecoder) DecodeMessage(msg []byte) (*exchange.EncryptedMessage, error) {
	var em pb.EncryptedMessage
	if err := proto.Unmarshal(msg, &em); err != nil {
		return nil, fmt.Errorf("protobufencode.DecodeMessage: %w", err)
	}
	return &exchange.EncryptedMessage{
		Type:    em.GetType(),
		KeyID:   em.GetKeyId(),
		Payload: em.GetPayload(),
	}, nil
}

func (EncoderDecoder) DecodeTicket(bs []byte) (*thirdparty.Ticket, error) {
	var mt pb.Ticket
	if err := proto.Unmarshal(bs, &mt); err != nil {
		return nil, fmt.Errorf("protobufencode.DecodeTicket: %w", err)
	}
	return &thirdparty.Ticket{
		CaveatKey: mt.GetKey(),
		Predicate: mt.GetPredicate(),
	}, nil
}
