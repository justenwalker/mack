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

func TestCaveatIDIssuer_IssueCaveatID(t *testing.T) {
	testErr := errors.New("test error")
	ticket := thirdparty.Ticket{
		CaveatKey: []byte(`12345678901234567890123456789012`),
		Predicate: []byte(`hello, world`),
	}
	tests := []struct {
		name      string
		expectErr error
		expectCID []byte
		setup     func(encoder *EncoderMock, encryptor *EncryptorMock)
	}{
		{
			name:      "err-encode-ticket",
			expectErr: testErr,
			setup: func(encoder *EncoderMock, _ *EncryptorMock) {
				encoder.EncodeTicketFunc = func(thirdparty.Ticket) ([]byte, error) {
					return nil, testErr
				}
			},
		},
		{
			name:      "err-encrypt-ticket",
			expectErr: testErr,
			setup: func(encoder *EncoderMock, encryptor *EncryptorMock) {
				encoder.EncodeTicketFunc = func(t thirdparty.Ticket) ([]byte, error) {
					if matchTicket(ticket).Matches(t) {
						return []byte(`encoded`), nil
					}
					panic(fmt.Errorf("unexpected message: %#v", t))
				}
				encryptor.EncryptMessageFunc = func(m []byte) (*exchange.EncryptedMessage, error) {
					if matchBytes(`encoded`).Matches(m) {
						return nil, testErr
					}
					panic(fmt.Errorf("unexpected message: %s", string(m)))
				}
			},
		},
		{
			name:      "err-encode-message",
			expectErr: testErr,
			setup: func(encoder *EncoderMock, encryptor *EncryptorMock) {
				encoder.EncodeTicketFunc = func(t thirdparty.Ticket) ([]byte, error) {
					if matchTicket(ticket).Matches(t) {
						return []byte(`encoded`), nil
					}
					panic(fmt.Errorf("unexpected message: %#v", t))
				}
				encryptor.EncryptMessageFunc = func(m []byte) (*exchange.EncryptedMessage, error) {
					if matchBytes(`encoded`).Matches(m) {
						return &exchange.EncryptedMessage{
							Type:    "none",
							KeyID:   "na",
							Payload: []byte(`encoded`),
						}, nil
					}
					panic(fmt.Errorf("unexpected message: %s", string(m)))
				}
				encoder.EncodeMessageFunc = func(*exchange.EncryptedMessage) ([]byte, error) {
					return nil, testErr
				}
			},
		},
		{
			name: "success",
			setup: func(encoder *EncoderMock, encryptor *EncryptorMock) {
				encoder.EncodeTicketFunc = func(t thirdparty.Ticket) ([]byte, error) {
					if matchTicket(ticket).Matches(t) {
						return []byte(`encoded`), nil
					}
					panic(fmt.Errorf("unexpected message: %#v", t))
				}
				encryptor.EncryptMessageFunc = func(m []byte) (*exchange.EncryptedMessage, error) {
					if matchBytes(`encoded`).Matches(m) {
						return &exchange.EncryptedMessage{
							Type:    "none",
							KeyID:   "na",
							Payload: []byte(`encoded`),
						}, nil
					}
					panic(fmt.Errorf("unexpected message: %s", string(m)))
				}
				encoder.EncodeMessageFunc = func(*exchange.EncryptedMessage) ([]byte, error) {
					return []byte(`encoded-message`), nil
				}
			},
			expectCID: []byte(`encoded-message`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var encoder EncoderMock
			var encryptor EncryptorMock
			tt.setup(&encoder, &encryptor)
			iss := exchange.CaveatIDIssuer{
				Encryptor: &encryptor,
				Encoder:   &encoder,
			}
			cID, err := iss.IssueCaveatID(ctx, ticket)
			if tt.expectErr != nil {
				if !errors.Is(err, tt.expectErr) {
					t.Fatalf("want %v, got %v", tt.expectErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.expectCID, cID, cmp.Comparer(bytesEqual)); diff != "" {
				t.Fatalf("IssueCaveatID unexpected cID (-want +got):\n%s", diff)
			}
		})
	}
}
