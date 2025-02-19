package libmacaroon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/justenwalker/mack"
	"github.com/justenwalker/mack/encoding"
)

var _ encoding.EncoderDecoder = V2{}

type V2 struct {
	OutputEncoder OutputEncoder
	InputDecoder  InputDecoder
}

func (v V2) String() string {
	return "libmacaroon/v2"
}

// DecodeMacaroon decodes a macaroon from libmacaroon v2 binary format.
func (v V2) DecodeMacaroon(bs []byte, m *mack.Macaroon) error {
	buf, err := decodeBuffer(v.InputDecoder, bs)
	if err != nil {
		return err
	}
	dec := NewV2Decoder(buf)
	return dec.DecodeMacaroon(m)
}

// DecodeStack decodes a stack of macaroons from libmacaroon v2 binary format.
func (v V2) DecodeStack(bs []byte, stack *mack.Stack) error {
	buf, err := decodeBuffer(v.InputDecoder, bs)
	if err != nil {
		return err
	}
	dec := NewV2Decoder(buf)
	return dec.DecodeStack(stack)
}

// EncodeMacaroon encodes a macaroon into libmacaroon v2 binary format.
func (v V2) EncodeMacaroon(m *mack.Macaroon) ([]byte, error) {
	sz := v2RawSizeBytes(m)
	if v.OutputEncoder != nil {
		sz = v.OutputEncoder.EncodedLength(sz)
	}
	buf := bytes.NewBuffer(make([]byte, 0, sz))
	var writer io.Writer = buf
	if v.OutputEncoder != nil {
		writer = v.OutputEncoder.EncodeOutput(writer)
	}
	enc := V2Encoder{writer: &byteWriter{Writer: writer}}
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

// EncodeStack encodes a stack of macaroons into libmacaroon v2 binary format.
func (v V2) EncodeStack(stack mack.Stack) ([]byte, error) {
	var sz int
	for i := range stack {
		sz += v2RawSizeBytes(&stack[i])
	}
	if v.OutputEncoder != nil {
		sz = v.OutputEncoder.EncodedLength(sz)
	}
	buf := bytes.NewBuffer(make([]byte, 0, sz))
	var writer io.Writer = buf
	if v.OutputEncoder != nil {
		writer = v.OutputEncoder.EncodeOutput(writer)
	}
	enc := NewV2Encoder(writer)
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
	v2VersionByte = byte(0x02)
)

type v2FieldType byte

const (
	v2FieldTypeEOS      v2FieldType = 0
	v2FieldTypeLocation v2FieldType = 1
	v2FieldTypeID       v2FieldType = 2
	v2FieldTypeVID      v2FieldType = 4
	v2FieldTypeSig      v2FieldType = 6
)

type V2Encoder struct {
	writer *byteWriter
}

func NewV2Encoder(w io.Writer) *V2Encoder {
	return &V2Encoder{writer: &byteWriter{Writer: w}}
}

func (enc *V2Encoder) EncodeMacaroon(m *mack.Macaroon) error {
	if err := enc.writer.WriteByte(v2VersionByte); err != nil {
		return err
	}
	if loc := m.Location(); loc != "" {
		if err := enc.writer.WriteByte(byte(v2FieldTypeLocation)); err != nil {
			return err
		}
		if err := enc.writeFieldString(loc); err != nil {
			return err
		}
	}
	if err := enc.writer.WriteByte(byte(v2FieldTypeID)); err != nil {
		return err
	}
	if err := enc.writeFieldValue(m.ID()); err != nil {
		return err
	}
	if err := enc.writer.WriteByte(byte(v2FieldTypeEOS)); err != nil {
		return err
	}
	cs := m.Caveats()
	for i := range cs {
		if err := enc.encodeCaveat(&cs[i]); err != nil {
			return err
		}
	}
	if err := enc.writer.WriteByte(byte(v2FieldTypeEOS)); err != nil {
		return err
	}
	if err := enc.writer.WriteByte(byte(v2FieldTypeSig)); err != nil {
		return err
	}
	if err := enc.writeFieldValue(m.Signature()); err != nil {
		return err
	}
	return nil
}

func (enc *V2Encoder) EncodeStack(stack mack.Stack) error {
	for i := range stack {
		if err := enc.EncodeMacaroon(&stack[i]); err != nil {
			return err
		}
	}
	return nil
}

func (enc *V2Encoder) encodeCaveat(c *mack.Caveat) error {
	if c.Location() != "" {
		if err := enc.writer.WriteByte(byte(v2FieldTypeLocation)); err != nil {
			return err
		}
		if err := enc.writeFieldString(c.Location()); err != nil {
			return err
		}
	}
	if err := enc.writer.WriteByte(byte(v2FieldTypeID)); err != nil {
		return err
	}
	if err := enc.writeFieldValue(c.ID()); err != nil {
		return err
	}
	vid := c.VID()
	if len(vid) > 0 {
		if err := enc.writer.WriteByte(byte(v2FieldTypeVID)); err != nil {
			return err
		}
		if err := enc.writeFieldValue(vid); err != nil {
			return err
		}
	}
	if err := enc.writer.WriteByte(byte(v2FieldTypeEOS)); err != nil {
		return err
	}
	return nil
}

func (ft v2FieldType) String() string {
	switch ft {
	case v2FieldTypeEOS:
		return "EOS"
	case v2FieldTypeLocation:
		return "location"
	case v2FieldTypeID:
		return "id"
	case v2FieldTypeVID:
		return "vid"
	case v2FieldTypeSig:
		return "sig"
	default:
		return fmt.Sprintf("v2FieldType(%x)", byte(ft))
	}
}

type V2Decoder struct {
	reader *byteReader
}

func NewV2Decoder(bs []byte) *V2Decoder {
	return &V2Decoder{reader: &byteReader{buf: bs}}
}

func (dec *V2Decoder) DecodeMacaroon(m *mack.Macaroon) error {
	ver, err := dec.reader.ReadByte()
	if err != nil {
		return fmt.Errorf("v2.DecodeMacaroon: could not read version byte: %w", err)
	}
	if ver != v2VersionByte {
		return fmt.Errorf("v2.DecodeMacaroon: invalid version byte: %x, expected=%x", ver, v2VersionByte)
	}
	var raw mack.Raw

	var (
		field v2FieldType
		data  []byte
		ok    bool
	)

	// Read Header Section
	for {
		ok, err = dec.readHeader(&raw)
		if err != nil {
			return fmt.Errorf("v2.DecodeMacaroon: could not read caveat: %w", err)
		}
		if !ok {
			break
		}
	}

	// Read Caveats Section
	for {
		ok, err = dec.readCaveat(&raw)
		if err != nil {
			return fmt.Errorf("v2.DecodeMacaroon: could not read caveat: %w", err)
		}
		if !ok {
			break
		}
	}

	// Read Signature
	field, data, err = dec.readField()
	if err != nil {
		return fmt.Errorf("v2.DecodeMacaroon: could not read signature field: %w", err)
	}
	if field != v2FieldTypeSig {
		return fmt.Errorf("v2.DecodeMacaroon: unexpected field type: %x", field)
	}
	raw.Signature = data
	*m = mack.NewFromRaw(raw)
	return nil
}

func (dec *V2Decoder) DecodeStack(stack *mack.Stack) error {
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

func (dec *V2Decoder) readHeader(m *mack.Raw) (bool, error) {
	field, data, err := dec.readField()
	if err != nil {
		return false, fmt.Errorf("v2.DecodeMacaroon: could not read field: %w", err)
	}
	switch field { //nolint:exhaustive
	case v2FieldTypeLocation:
		if len(m.ID) > 0 {
			return false, errors.New("v2.DecodeMacaroon: 'location' encountered after 'id'")
		}
		if m.Location != "" {
			return false, errors.New("v2.DecodeMacaroon: duplicate field 'location'")
		}
		m.Location = string(data)
		return true, nil
	case v2FieldTypeID:
		if len(m.ID) != 0 {
			return false, errors.New("v2.DecodeMacaroon: duplicate field 'id'")
		}
		m.ID = data
		return true, nil
	case v2FieldTypeEOS:
		return false, nil // no caveat
	default:
		return false, fmt.Errorf("v2.DecodeMacaroon: unexpected field type: %x", field)
	}
}

func (dec *V2Decoder) readCaveat(m *mack.Raw) (bool, error) {
	var c mack.RawCaveat
	field, data, err := dec.readField()
	if err != nil {
		return false, fmt.Errorf("v2.DecodeMacaroon: could not read caveat field: %w", err)
	}
	switch field { //nolint:exhaustive
	case v2FieldTypeLocation:
		c.Location = string(data)
	case v2FieldTypeID:
		c.CID = data
	case v2FieldTypeEOS:
		return false, nil // no caveat
	default:
		return false, fmt.Errorf("v2.DecodeMacaroon: unexpected field type: %x", field)
	}

	if len(c.CID) == 0 { // ensure we get the CID
		field, data, err = dec.readField()
		if err != nil {
			return false, fmt.Errorf("v2.DecodeMacaroon: could not read caveat field: %w", err)
		}
		if field != v2FieldTypeID {
			return false, fmt.Errorf("v2.DecodeMacaroon: unexpected caveat field type: %x", field)
		}
		c.CID = data
	}

	field, data, err = dec.readField()
	if err != nil {
		return false, fmt.Errorf("v2.DecodeMacaroon: could not read caveat field: %w", err)
	}
	switch field { //nolint:exhaustive
	case v2FieldTypeVID:
		c.VID = data
	case v2FieldTypeEOS: // optional VID not given
		m.Caveats = append(m.Caveats, c)
		return true, nil
	default:
		return false, fmt.Errorf("v2.DecodeMacaroon: unexpected caveat field type: %x", field)
	}

	// expect EOS
	field, _, err = dec.readField()
	if err != nil {
		return false, fmt.Errorf("v2.DecodeMacaroon: could not read caveat field: %w", err)
	}
	if field != v2FieldTypeEOS {
		return false, fmt.Errorf("v2.DecodeMacaroon: unexpected caveat field type: %x", field)
	}
	m.Caveats = append(m.Caveats, c)
	return true, nil
}

func (dec *V2Decoder) readField() (v2FieldType, []byte, error) {
	b, err := dec.reader.ReadByte()
	if err != nil {
		return 0, nil, fmt.Errorf("fail to read field type: %w", err)
	}
	ft := v2FieldType(b)
	if ft == v2FieldTypeEOS {
		return ft, nil, nil
	}
	fieldLen, err := binary.ReadUvarint(dec.reader)
	if err != nil {
		return 0, nil, fmt.Errorf("fail to read field len: %w", err)
	}
	value, err := dec.reader.ReadField(int(fieldLen))
	if err != nil {
		return 0, nil, fmt.Errorf("fail to read field data (size=%d): %w", fieldLen, err)
	}
	return ft, value, nil
}

func (enc *V2Encoder) writeFieldValue(data []byte) error {
	if _, err := enc.writer.WriteVarint(uint64(len(data))); err != nil {
		return err
	}
	if _, err := enc.writer.Write(data); err != nil {
		return err
	}
	return nil
}

func (enc *V2Encoder) writeFieldString(str string) error {
	if _, err := enc.writer.WriteVarint(uint64(len(str))); err != nil {
		return err
	}
	if _, err := enc.writer.WriteString(str); err != nil {
		return err
	}
	return nil
}

func v2RawSizeBytes(m *mack.Macaroon) int {
	var varint [8]byte
	var n int
	sz := 1 // version byte

	// location
	if n = len(m.Location()); n > 0 {
		sz += 1 + binary.PutUvarint(varint[:], uint64(n)) + n
	}
	// id
	n = len(m.ID())
	sz += 1 + binary.PutUvarint(varint[:], uint64(n)) + n
	sz++ // eos

	for _, c := range m.Caveats() {
		// cl
		if n = len(c.Location()); n > 0 {
			sz += 1 + binary.PutUvarint(varint[:], uint64(n)) + n
		}
		// id
		n = len(c.ID())
		sz += 1 + binary.PutUvarint(varint[:], uint64(n)) + n

		// vid
		if n = len(c.VID()); n > 0 {
			sz += 1 + binary.PutUvarint(varint[:], uint64(n)) + n
		}
		sz++ // eos
	}
	sz++ // eos

	// sig
	n = len(m.Signature())
	sz += 1 + binary.PutUvarint(varint[:], uint64(n)) + n
	return sz
}
