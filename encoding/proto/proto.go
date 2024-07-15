// Package proto provides an Encoding for various macaroon types using protobuf wire format.
package proto

// Encoding implements the macaroon.Encoder, macaroon.Decoder, exchange.Encoder,
// and exchange.Decoder interfaces. It provides methods to encode and decode different types of structured
// data using protobuf wire format.
var Encoding EncoderDecoder

// EncoderDecoder is a type that implements the macaroon.Encoder, macaroon.Decoder, exchange.Encoder,
// and exchange.Decoder interfaces. It provides methods to encode and decode different types of structured
// data using protobuf wire format.
type EncoderDecoder struct{}
