package proto

import (
	"google.golang.org/protobuf/proto"

	"github.com/justenwalker/mack/encoding"
	"github.com/justenwalker/mack/encoding/proto/pb"
	"github.com/justenwalker/mack/macaroon"
)

var (
	_ encoding.MacaroonEncoder = EncoderDecoder{}
	_ encoding.MacaroonDecoder = EncoderDecoder{}
	_ encoding.StackEncoder    = EncoderDecoder{}
	_ encoding.StackDecoder    = EncoderDecoder{}
)

// EncodeStack encodes the Macaroon authorization Stack into bytes using protobuf wire format.
func (EncoderDecoder) EncodeStack(stack macaroon.Stack) ([]byte, error) {
	var req pb.Stack
	req.Macaroons = make([]*pb.Macaroon, len(stack))
	for i := range stack {
		req.Macaroons[i] = macaroonToProto(&stack[i])
	}
	return proto.Marshal(&req)
}

// DecodeStack decodes the protobuf bytes into the target Macaroon authorization Stack.
func (EncoderDecoder) DecodeStack(bs []byte, stack *macaroon.Stack) (err error) {
	var req pb.Stack
	if err = proto.Unmarshal(bs, &req); err != nil {
		return err
	}
	s := make(macaroon.Stack, 0, len(req.GetMacaroons()))
	for _, m := range req.GetMacaroons() {
		s = append(s, protoToMacaroon(m))
	}
	*stack = s
	return nil
}

// EncodeMacaroon encodes the macaroon using protobuf into the returned byte slice, with no additional header.
func (EncoderDecoder) EncodeMacaroon(m *macaroon.Macaroon) ([]byte, error) {
	return proto.Marshal(macaroonToProto(m))
}

func macaroonToProto(m *macaroon.Macaroon) *pb.Macaroon {
	var caveats []*pb.Caveat
	cs := m.Caveats()
	if len(cs) > 0 {
		caveats = make([]*pb.Caveat, len(cs))
		for i := range cs {
			caveats[i] = &pb.Caveat{
				Cid: cs[i].ID(),
				Vid: cs[i].VID(),
				Cl:  cs[i].Location(),
			}
		}
	}
	return &pb.Macaroon{
		Loc:     m.Location(),
		Id:      m.ID(),
		Caveats: caveats,
		Sig:     m.Signature(),
	}
}

// DecodeMacaroon decodes the given byte slice into the macaroon provided.
// The byte slice is expected to be a raw protobuf, with no additional header.
func (EncoderDecoder) DecodeMacaroon(bs []byte, m *macaroon.Macaroon) error {
	var p pb.Macaroon
	if err := proto.Unmarshal(bs, &p); err != nil {
		return err
	}
	*m = protoToMacaroon(&p)
	return nil
}

func protoToMacaroon(p *pb.Macaroon) macaroon.Macaroon {
	caveats := make([]macaroon.RawCaveat, len(p.GetCaveats()))
	for i := range p.GetCaveats() {
		caveats[i] = macaroon.RawCaveat{
			CID:      p.GetCaveats()[i].GetCid(),
			VID:      p.GetCaveats()[i].GetVid(),
			Location: p.GetCaveats()[i].GetCl(),
		}
	}
	return macaroon.NewFromRaw(macaroon.Raw{
		Location:  p.GetLoc(),
		ID:        p.GetId(),
		Caveats:   caveats,
		Signature: p.GetSig(),
	})
}
