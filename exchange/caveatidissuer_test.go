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
		setup     func(encoder *MockEncoder, encryptor *MockEncryptor)
	}{
		{
			name:      "err-encode-ticket",
			expectErr: testErr,
			setup: func(encoder *MockEncoder, _ *MockEncryptor) {
				encoder.EXPECT().EncodeTicket(gomock.Any()).Return(nil, testErr)
			},
		},
		{
			name:      "err-encrypt-ticket",
			expectErr: testErr,
			setup: func(encoder *MockEncoder, encryptor *MockEncryptor) {
				encoder.EXPECT().EncodeTicket(matchTicket(ticket)).Return([]byte(`encoded`), nil)
				encryptor.EXPECT().EncryptMessage(matchBytes(`encoded`)).Return(nil, testErr)
			},
		},
		{
			name:      "err-encode-message",
			expectErr: testErr,
			setup: func(encoder *MockEncoder, encryptor *MockEncryptor) {
				encoder.EXPECT().EncodeTicket(matchTicket(ticket)).Return([]byte(`encoded`), nil)
				encryptor.EXPECT().EncryptMessage(matchBytes(`encoded`)).Return(&exchange.EncryptedMessage{
					Type:    "none",
					KeyID:   "na",
					Payload: []byte(`encoded`),
				}, nil)
				encoder.EXPECT().EncodeMessage(gomock.Any()).Return(nil, testErr)
			},
		},
		{
			name: "success",
			setup: func(encoder *MockEncoder, encryptor *MockEncryptor) {
				encoder.EXPECT().EncodeTicket(matchTicket(ticket)).Return([]byte(`encoded`), nil)
				encryptor.EXPECT().EncryptMessage(matchBytes(`encoded`)).Return(&exchange.EncryptedMessage{
					Type:    "none",
					KeyID:   "na",
					Payload: []byte(`encoded`),
				}, nil)
				encoder.EXPECT().EncodeMessage(gomock.Any()).Return([]byte(`encoded-message`), nil)
			},
			expectCID: []byte(`encoded-message`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			encoder := NewMockEncoder(ctrl)
			encryptor := NewMockEncryptor(ctrl)
			tt.setup(encoder, encryptor)
			iss := exchange.CaveatIDIssuer{
				Encryptor: encryptor,
				Encoder:   encoder,
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
