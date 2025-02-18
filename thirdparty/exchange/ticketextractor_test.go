package exchange_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/justenwalker/mack/thirdparty"
	"github.com/justenwalker/mack/thirdparty/exchange"
)

func TestTicketExtractor_ExtractTicket(t *testing.T) {
	testErr := errors.New("test error")
	cID := []byte(`caveat-identifier`)
	ticket := thirdparty.Ticket{
		CaveatKey: []byte(`12345678901234567890123456789012`),
		Predicate: []byte(`hello, world`),
	}
	encryptedMessage := exchange.EncryptedMessage{
		Type:    "none",
		KeyID:   "na",
		Payload: []byte(`encrypted`),
	}
	tests := []struct {
		name         string
		expectErr    error
		expectTicket thirdparty.Ticket
		setup        func(decoder *DecoderMock, decryptor *DecryptorMock)
	}{
		{
			name:      "err-decode-message",
			expectErr: testErr,
			setup: func(decoder *DecoderMock, _ *DecryptorMock) {
				decoder.DecodeMessageFunc = func(c []byte) (*exchange.EncryptedMessage, error) {
					if matchBytes(cID).Matches(c) {
						return nil, testErr
					}
					panic(fmt.Errorf("unexpected cID: %s", string(c)))
				}
			},
		},
		{
			name:      "err-decrypt-message",
			expectErr: testErr,
			setup: func(decoder *DecoderMock, decryptor *DecryptorMock) {
				decoder.DecodeMessageFunc = func(c []byte) (*exchange.EncryptedMessage, error) {
					if matchBytes(cID).Matches(c) {
						return &encryptedMessage, nil
					}
					panic(fmt.Errorf("unexpected cID: %s", string(c)))
				}
				decryptor.DecryptMessageFunc = func(em *exchange.EncryptedMessage) ([]byte, error) {
					if matchEncryptedMessage(encryptedMessage).Matches(em) {
						return nil, testErr
					}
					panic(fmt.Errorf("unexpected EncryptedMessage: %v", em))
				}
			},
		},
		{
			name:      "err-decode-ticket",
			expectErr: testErr,
			setup: func(decoder *DecoderMock, decryptor *DecryptorMock) {
				decoder.DecodeMessageFunc = func(c []byte) (*exchange.EncryptedMessage, error) {
					if matchBytes(cID).Matches(c) {
						return &encryptedMessage, nil
					}
					panic(fmt.Errorf("unexpected cID: %s", string(c)))
				}
				decryptor.DecryptMessageFunc = func(em *exchange.EncryptedMessage) ([]byte, error) {
					if matchEncryptedMessage(encryptedMessage).Matches(em) {
						return []byte(`decrypted`), nil
					}
					panic(fmt.Errorf("unexpected EncryptedMessage: %v", em))
				}
				decoder.DecodeTicketFunc = func(bs []byte) (*thirdparty.Ticket, error) {
					if matchBytes(`decrypted`).Matches(bs) {
						return nil, testErr
					}
					panic(fmt.Errorf("unexpected message: %s", string(bs)))
				}
			},
		},
		{
			name: "success",
			setup: func(decoder *DecoderMock, decryptor *DecryptorMock) {
				decoder.DecodeMessageFunc = func(c []byte) (*exchange.EncryptedMessage, error) {
					if matchBytes(cID).Matches(c) {
						return &encryptedMessage, nil
					}
					panic(fmt.Errorf("unexpected cID: %s", string(c)))
				}
				decryptor.DecryptMessageFunc = func(em *exchange.EncryptedMessage) ([]byte, error) {
					if matchEncryptedMessage(encryptedMessage).Matches(em) {
						return []byte(`decrypted`), nil
					}
					panic(fmt.Errorf("unexpected EncryptedMessage: %v", em))
				}
				decoder.DecodeTicketFunc = func(bs []byte) (*thirdparty.Ticket, error) {
					if matchBytes(`decrypted`).Matches(bs) {
						return &ticket, nil
					}
					panic(fmt.Errorf("unexpected message: %s", string(bs)))
				}
			},
			expectTicket: ticket,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var decoder DecoderMock
			var decryptor DecryptorMock
			tt.setup(&decoder, &decryptor)
			text := exchange.TicketExtractor{
				Decryptor: &decryptor,
				Decoder:   &decoder,
			}
			tk, err := text.ExtractTicket(ctx, cID)
			if tt.expectErr != nil {
				if !errors.Is(err, tt.expectErr) {
					t.Fatalf("want %v, got %v", tt.expectErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.expectTicket, *tk, cmp.Comparer(ticketsEqual)); diff != "" {
				t.Fatalf("ExtractTicket unexpected ticket (-want +got):\n%s", diff)
			}
		})
	}
}
