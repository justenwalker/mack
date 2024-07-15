// Package exchange provides implementation for the thirdparty.CaveatIDIssuer and thirdparty.TicketExtractor.
//
// CaveatIDIssuer encrypts a thirdparty.Ticket into an EncryptedMessage using an Encryptor,
// and encoding it into bytes via an Encoder. The result is used as the CaveatID directly.
//
// TicketExtractor extracts the thirdparty.Ticket from the CaveatID by doing the reverse of CaveatIDIssuer
// via the corresponding Decoder and Decryptor.
package exchange

import "github.com/justenwalker/mack/macaroon/thirdparty"

// Encoder is an interface that encodes structured messages into bytes.
// These raw bytes can be later decoded by a Decoder.
type Encoder interface {
	// EncodeMessage is a method of the Encoder interface that encodes an EncryptedMessage
	// into a byte array. It returns the encoded message as a byte slice and any error encountered during encoding.
	EncodeMessage(em *EncryptedMessage) ([]byte, error)

	// EncodeTicket is a method of the Encoder interface that encodes a thirdparty.Ticket into a byte slice.
	EncodeTicket(t thirdparty.Ticket) ([]byte, error)
}

// Encryptor encrypts bytes into an encapsulated EncryptedMessage structure.
// which can be later decrypted by a Decryptor.
type Encryptor interface {
	EncryptMessage(msg []byte) (*EncryptedMessage, error)
}

// TicketExtractor implements thirdparty.TicketExtractor with arbitrary Decoder and Decryptor implementations.
// This implementation is the other side of the CaveatIDIssuer which extracts the Ticket from the caveat id (cId).
type TicketExtractor struct {
	// Decryptor decrypts the EncryptedMessage.Payload (Required).
	Decryptor Decryptor
	// Decoder decodes tickets from the caveat identifier.
	Decoder Decoder
}

// Decoder is an interface that provides methods to decode messages and tickets.
// It decodes a byte array into an EncryptedMessage struct or a thirdparty.Ticket struct.
type Decoder interface {
	// DecodeMessage decodes a byte array into an EncryptedMessage struct.
	DecodeMessage(msg []byte) (*EncryptedMessage, error)

	// DecodeTicket decodes a byte array into a thirdparty.Ticket struct.
	DecodeTicket(bs []byte) (*thirdparty.Ticket, error)
}

// Decryptor is an interface that provides a method to decrypt an EncryptedMessage.
type Decryptor interface {
	// DecryptMessage decrypts an EncryptedMessage its decrypted payload as a plain byte array.
	DecryptMessage(em *EncryptedMessage) ([]byte, error)
}
