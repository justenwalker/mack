package exchange_test

import (
	"bytes"
	"fmt"

	"go.uber.org/mock/gomock"

	"github.com/justenwalker/mack/exchange"
	"github.com/justenwalker/mack/macaroon/thirdparty"
)

//go:generate go run go.uber.org/mock/mockgen -source exchange.go -destination exchange_mock_test.go -package exchange_test

type matchEncryptedMessage exchange.EncryptedMessage

func (mem matchEncryptedMessage) Matches(x any) bool {
	var o exchange.EncryptedMessage
	switch t := x.(type) {
	case exchange.EncryptedMessage:
		o = t
	case *exchange.EncryptedMessage:
		if t == nil {
			return false
		}
		o = *t
	default:
		return false
	}
	if o.Type != mem.Type {
		return false
	}
	if o.KeyID != mem.KeyID {
		return false
	}
	return bytesEqual(o.Payload, mem.Payload)
}

func (mem matchEncryptedMessage) String() string {
	return fmt.Sprintf("%#v", exchange.EncryptedMessage(mem))
}

type matchTicket thirdparty.Ticket

func (mt matchTicket) Matches(x any) bool {
	var o thirdparty.Ticket
	switch t := x.(type) {
	case thirdparty.Ticket:
		o = t
	case *thirdparty.Ticket:
		if t == nil {
			return false
		}
		o = *t
	default:
		return false
	}
	return ticketsEqual(o, thirdparty.Ticket(mt))
}

func ticketsEqual(a thirdparty.Ticket, b thirdparty.Ticket) bool {
	if !bytesEqual(a.CaveatKey, b.CaveatKey) {
		return false
	}
	return bytesEqual(a.Predicate, b.Predicate)
}

func (mt matchTicket) String() string {
	return fmt.Sprintf("%#v", thirdparty.Ticket(mt))
}

type matchBytes []byte

func (mb matchBytes) Matches(x any) bool {
	if bs, ok := x.([]byte); ok {
		return bytesEqual(mb, bs)
	}
	return false
}

func (mb matchBytes) String() string {
	return fmt.Sprintf("[]byte(%x)", []byte(mb))
}

var _ gomock.Matcher = &matchBytes{}

func bytesEqual(a []byte, b []byte) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	return bytes.Equal(a, b)
}
