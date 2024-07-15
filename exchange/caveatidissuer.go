package exchange

import (
	"context"
	"fmt"

	"github.com/justenwalker/mack/macaroon/thirdparty"
)

var _ thirdparty.CaveatIDIssuer = (*CaveatIDIssuer)(nil)

// CaveatIDIssuer implements thirdparty.CaveatIDIssuer with arbitrary Encoder and Encryptor implementations.
// This is used to produce a caveat id from a thirdparty.Ticket that can be extracted later from a TicketExtractor.
type CaveatIDIssuer struct {
	// Encryptor encrypts the caveat key and predicate used to generate a Discharge Macaroon.
	Encryptor Encryptor
	// Encoder encodes the encrypted message into the caveat ID bytes.
	Encoder Encoder
}

// EncryptedMessage is a structure that represents an encrypted message.
// It contains the type of the message, the key ID, and the payload.
type EncryptedMessage struct {
	// Type is a hint about the type of encryptor that encrypted this message
	Type string
	// KeyID is a hint about which key was used to encrypt the payload
	KeyID string
	// Payload is the encrypted message bytes.
	Payload []byte
}

func (e *CaveatIDIssuer) IssueCaveatID(_ context.Context, t thirdparty.Ticket) (cID []byte, err error) {
	plain, err := e.Encoder.EncodeTicket(t)
	if err != nil {
		return nil, fmt.Errorf("EncodeTicket: %w", err)
	}
	encMsg, err := e.Encryptor.EncryptMessage(plain)
	if err != nil {
		return nil, fmt.Errorf("EncryptMessage: %w", err)
	}
	cID, err = e.Encoder.EncodeMessage(encMsg)
	if err != nil {
		return nil, fmt.Errorf("EncodeMessage: %w", err)
	}
	return cID, nil
}
