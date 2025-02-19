package libmacaroon

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"unsafe"

	"github.com/justenwalker/mack"
	"github.com/justenwalker/mack/encoding"
)

var _ encoding.EncoderDecoder = V1{}

type V1 struct {
	OutputEncoder OutputEncoder
	InputDecoder  InputDecoder
}

func (v V1) String() string {
	return "libmacaroon/v1"
}

// DecodeMacaroon decodes a macaroon from libmacaroon v1 binary format.
func (v V1) DecodeMacaroon(bs []byte, m *mack.Macaroon) error {
	buf, err := decodeBuffer(v.InputDecoder, bs)
	if err != nil {
		return err
	}
	dec := NewV1Decoder(buf)
	return dec.DecodeMacaroon(m)
}

// DecodeStack decodes a stack of macaroons from libmacaroon v1 binary format.
func (v V1) DecodeStack(bs []byte, stack *mack.Stack) error {
	buf, err := decodeBuffer(v.InputDecoder, bs)
	if err != nil {
		return err
	}
	dec := NewV1Decoder(buf)
	return dec.DecodeStack(stack)
}

// EncodeMacaroon encodes a macaroon into libmacaroon v1 binary format.
func (v V1) EncodeMacaroon(m *mack.Macaroon) ([]byte, error) {
	sz := v1RawSizeBytes(m)
	if v.OutputEncoder != nil {
		sz = v.OutputEncoder.EncodedLength(sz)
	}
	sz = base64.URLEncoding.EncodedLen(sz)
	buf := bytes.NewBuffer(make([]byte, 0, sz))
	var writer io.Writer = buf
	if v.OutputEncoder != nil {
		writer = v.OutputEncoder.EncodeOutput(writer)
	}
	enc := V1Encoder{writer: &byteWriter{Writer: writer}}
	if err := enc.EncodeMacaroon(m); err != nil {
		return nil, err
	}
	if wc, ok := writer.(io.Closer); ok {
		if err := wc.Close(); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// EncodeStack encodes a stack of macaroons into libmacaroon v1 binary format.
func (v V1) EncodeStack(stack mack.Stack) ([]byte, error) {
	var buf bytes.Buffer
	var writer io.Writer = &buf
	if v.OutputEncoder != nil {
		writer = v.OutputEncoder.EncodeOutput(writer)
	}
	enc := NewV1Encoder(writer)
	if err := enc.EncodeStack(stack); err != nil {
		return nil, err
	}
	if wc, ok := writer.(io.Closer); ok {
		if err := wc.Close(); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

const (
	v1FieldLocation       = v1FieldType("location")
	v1FieldIdentifier     = v1FieldType("identifier")
	v1FieldSignature      = v1FieldType("signature")
	v1FieldCid            = v1FieldType("cid")
	v1FieldVerification   = v1FieldType("vid")
	v1FieldCaveatLocation = v1FieldType("cl")
)

type V1Encoder struct {
	writer io.Writer
}

func NewV1Encoder(w io.Writer) *V1Encoder {
	return &V1Encoder{writer: &byteWriter{Writer: w}}
}

func (enc *V1Encoder) EncodeMacaroon(m *mack.Macaroon) error {
	bw := byteWriter{Writer: enc.writer}
	if err := v1WriteLocation(&bw, m.Location()); err != nil {
		return fmt.Errorf("v1.Encoder: failed to write field '%s': %w", v1FieldLocation, err)
	}
	if err := v1WritePacket(&bw, v1FieldIdentifier, m.ID()); err != nil {
		return fmt.Errorf("v1.Encoder: failed to write field '%s': %w", v1FieldIdentifier, err)
	}
	cs := m.Caveats()
	for i := range cs {
		if err := v1WriteCaveat(&bw, &cs[i]); err != nil {
			return err
		}
	}
	if err := v1WritePacket(&bw, v1FieldSignature, m.Signature()); err != nil {
		return fmt.Errorf("v1.Encoder: failed to write field '%s': %w", v1FieldSignature, err)
	}
	return nil
}

func (enc *V1Encoder) EncodeStack(stack mack.Stack) error {
	for i := range stack {
		if err := enc.EncodeMacaroon(&stack[i]); err != nil {
			return err
		}
	}
	return nil
}

type V1Decoder struct {
	reader *byteReader
}

func NewV1Decoder(bs []byte) *V1Decoder {
	return &V1Decoder{reader: &byteReader{buf: bs}}
}

func (dec *V1Decoder) DecodeMacaroon(m *mack.Macaroon) error {
	var raw mack.Raw
	var (
		field v1FieldType
		data  []byte
		err   error
	)
	// Location
	field, data, err = v1ReadPacket(dec.reader)
	if err != nil {
		return fmt.Errorf("v1.DecodeMacaroon: could not read location field: %w", err)
	}
	if field != v1FieldLocation {
		return fmt.Errorf("v1.DecodeMacaroon: unexpected field '%s': %w", field, err)
	}
	raw.Location = string(data)

	// ID
	field, data, err = v1ReadPacket(dec.reader)
	if err != nil {
		return fmt.Errorf("v1.DecodeMacaroon: could not read identifier field: %w", err)
	}
	if field != v1FieldIdentifier {
		return fmt.Errorf("v1.DecodeMacaroon: unexpected field '%s': %w", field, err)
	}
	raw.ID = data

	var c mack.RawCaveat
	for {
		field, data, err = v1ReadPacket(dec.reader)
		if err != nil {
			return fmt.Errorf("v1.DecodeMacaroon: could not read caveat: %w", err)
		}
		switch field { //nolint:exhaustive
		case v1FieldCid:
			if len(c.CID) != 0 { // another caveat immediately after the last CID
				raw.Caveats = append(raw.Caveats, c)
				c = mack.RawCaveat{}
			}
			c.CID = data
		case v1FieldCaveatLocation:
			if c.Location != "" {
				return fmt.Errorf("v1.DecodeMacaroon: duplicate caveat field '%s': %w", field, err)
			}
			c.Location = string(data)
		case v1FieldVerification:
			if len(c.VID) != 0 {
				return fmt.Errorf("v1.DecodeMacaroon: duplicate caveat field '%s': %w", field, err)
			}
			c.VID = data
		case v1FieldSignature: // done with caveats, signature means we're at the end of the macaroon
			if len(c.CID) != 0 {
				raw.Caveats = append(raw.Caveats, c)
			}
			raw.Signature = data
			*m = mack.NewFromRaw(raw) // convert
			return nil
		default:
			return fmt.Errorf("v1.DecodeMacaroon: unexpected field '%s': %w", field, err)
		}
	}
}

func (dec *V1Decoder) DecodeStack(stack *mack.Stack) error {
	var s mack.Stack
	for {
		var m mack.Macaroon
		err := dec.DecodeMacaroon(&m)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		s = append(s, m)
	}
	*stack = s
	return nil
}

type v1FieldType string

func v1ReadPacket(r io.Reader) (v1FieldType, []byte, error) {
	var fieldLenBytes [4]byte
	if _, err := io.ReadFull(r, fieldLenBytes[:]); err != nil {
		return "", nil, fmt.Errorf("read field length: %w", err)
	}
	var fieldLenDecoded [2]byte
	if _, err := hex.Decode(fieldLenDecoded[:], fieldLenBytes[:]); err != nil {
		return "", nil, fmt.Errorf("read field length: %w", err)
	}
	fieldLen := binary.BigEndian.Uint16(fieldLenDecoded[:]) - 4 // remove size overhead
	fieldBytes := make([]byte, fieldLen)

	if _, err := io.ReadFull(r, fieldBytes); err != nil {
		return "", nil, fmt.Errorf("read field data: %w", err)
	}
	sp := bytes.IndexByte(fieldBytes, 0x20) // SP
	if sp == -1 {
		return "", nil, errors.New("field data should SP (0x20) separating key and value")
	}
	lf := len(fieldBytes)
	if fieldBytes[lf-1] != 0x0A { // LF
		return "", nil, errors.New("field data end with a LF character (0x0A)")
	}
	key := fieldBytes[:sp]
	value := fieldBytes[sp+1 : lf-1]
	return v1FieldType(key), value, nil
}

func v1WriteLocation(w *byteWriter, loc string) error {
	n := len(loc)
	bp := unsafe.StringData(loc)
	bs := unsafe.Slice(bp, n)
	return v1WritePacket(w, v1FieldLocation, bs)
}

func v1WritePacket(w *byteWriter, ft v1FieldType, data []byte) error {
	var lengthBytes [2]byte
	var lengthHex [4]byte
	binary.BigEndian.PutUint16(lengthBytes[:], safeUint16(len(data)+6+len(ft)))
	hex.Encode(lengthHex[:], lengthBytes[:])
	if _, err := w.Write(lengthHex[:]); err != nil { // Length
		return err
	}
	if _, err := w.WriteString(string(ft)); err != nil { // Field Type
		return err
	}
	if err := w.WriteByte(0x20); err != nil { // SPC
		return err
	}
	if _, err := w.Write(data); err != nil { // Data
		return err
	}
	if err := w.WriteByte(0x0A); err != nil { // LF
		return err
	}
	return nil
}

func v1WriteCaveat(bw *byteWriter, c *mack.Caveat) error {
	if err := v1WritePacket(bw, v1FieldCid, c.ID()); err != nil {
		return fmt.Errorf("v1.Encoder: failed to write caveat field '%s': %w", v1FieldCid, err)
	}
	if len(c.VID()) == 0 {
		return nil
	}
	if err := v1WritePacket(bw, v1FieldVerification, c.VID()); err != nil {
		return fmt.Errorf("v1.Encoder: failed to write caveat field '%s': %w", v1FieldVerification, err)
	}
	if err := v1WriteLocation(bw, c.Location()); err != nil {
		return fmt.Errorf("v1.Encoder: failed to write caveat field '%s': %w", v1FieldLocation, err)
	}
	return nil
}

func v1RawSizeBytes(m *mack.Macaroon) (sz int) {
	sz += 6 + len(v1FieldLocation) + len(m.Location())
	sz += 6 + len(v1FieldIdentifier) + len(m.ID())
	sz += 6 + len(v1FieldSignature) + len(m.Signature())
	for _, c := range m.Caveats() {
		sz += 6 + len(v1FieldCid) + len(c.ID())
		if vid := c.VID(); len(vid) > 0 {
			sz += 6 + len(v1FieldVerification) + len(vid)
			sz += 6 + len(v1FieldCaveatLocation) + len(c.Location())
		}
	}
	return sz
}

func safeUint16(n int) uint16 {
	if n > 65535 {
		return 65535
	}
	if n < 0 {
		return 0
	}
	return uint16(n)
}
