package exchange_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"github.com/justenwalker/mack/exchange"
	"github.com/justenwalker/mack/macaroon/thirdparty"
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
		setup        func(decoder *MockDecoder, decryptor *MockDecryptor)
	}{
		{
			name:      "err-decode-message",
			expectErr: testErr,
			setup: func(decoder *MockDecoder, _ *MockDecryptor) {
				decoder.EXPECT().DecodeMessage(matchBytes(cID)).Return(nil, testErr)
			},
		},
		{
			name:      "err-decrypt-message",
			expectErr: testErr,
			setup: func(decoder *MockDecoder, decryptor *MockDecryptor) {
				decoder.EXPECT().DecodeMessage(matchBytes(cID)).Return(&encryptedMessage, nil)
				decryptor.EXPECT().DecryptMessage(matchEncryptedMessage(encryptedMessage)).Return(nil, testErr)
			},
		},
		{
			name:      "err-decode-ticket",
			expectErr: testErr,
			setup: func(decoder *MockDecoder, decryptor *MockDecryptor) {
				decoder.EXPECT().DecodeMessage(matchBytes(cID)).Return(&encryptedMessage, nil)
				decryptor.EXPECT().DecryptMessage(matchEncryptedMessage(encryptedMessage)).Return([]byte(`decrypted`), nil)
				decoder.EXPECT().DecodeTicket(matchBytes(`decrypted`)).Return(nil, testErr)
			},
		},
		{
			name: "success",
			setup: func(decoder *MockDecoder, decryptor *MockDecryptor) {
				decoder.EXPECT().DecodeMessage(matchBytes(cID)).Return(&encryptedMessage, nil)
				decryptor.EXPECT().DecryptMessage(matchEncryptedMessage(encryptedMessage)).Return([]byte(`decrypted`), nil)
				decoder.EXPECT().DecodeTicket(matchBytes(`decrypted`)).Return(&ticket, nil)
			},
			expectTicket: ticket,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			decoder := NewMockDecoder(ctrl)
			decryptor := NewMockDecryptor(ctrl)
			tt.setup(decoder, decryptor)
			text := exchange.TicketExtractor{
				Decryptor: decryptor,
				Decoder:   decoder,
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
