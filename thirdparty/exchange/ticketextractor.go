package exchange

import (
	"context"
	"fmt"

	"github.com/justenwalker/mack/thirdparty"
)

var _ thirdparty.TicketExtractor = (*TicketExtractor)(nil)

// ExtractTicket extracts a ticket from the given caveat identifier.
// It decodes an encrypted message from the cID, decrypts it,
// and decodes the ticket from the decrypted payload.
// The cID bytes are generally produced by the CaveatIDIssuer in this package.
func (t *TicketExtractor) ExtractTicket(_ context.Context, cID []byte) (*thirdparty.Ticket, error) {
	encMsg, err := t.Decoder.DecodeMessage(cID)
	if err != nil {
		return nil, fmt.Errorf("DecodeMessage: %w", err)
	}
	decMsg, err := t.Decryptor.DecryptMessage(encMsg)
	if err != nil {
		return nil, fmt.Errorf("DecryptMessage: %w", err)
	}
	ticket, err := t.Decoder.DecodeTicket(decMsg)
	if err != nil {
		return nil, fmt.Errorf("DecodeTicket: %w", err)
	}
	return ticket, nil
}
