package libmacaroon

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/justenwalker/mack"
	"github.com/justenwalker/mack/encoding"
)

var _ encoding.EncoderDecoder = V1J{}

type V1J struct{}

func (V1J) String() string {
	return "libmacaroon/v1j"
}

// DecodeMacaroon decodes a macaroon from libmacaroon v1 json format.
func (V1J) DecodeMacaroon(bs []byte, m *mack.Macaroon) error {
	dec := NewV1JDecoder(bs)
	return dec.DecodeMacaroon(m)
}

// DecodeStack decodes a macaroon stack from v1 json format.
func (V1J) DecodeStack(bs []byte, stack *mack.Stack) error {
	dec := NewV1JDecoder(bs)
	return dec.DecodeStack(stack)
}

// EncodeMacaroon encodes a macaroon into libmacaroon v1 json format.
func (V1J) EncodeMacaroon(m *mack.Macaroon) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewV1JEncoder(&buf)
	if err := enc.EncodeMacaroon(m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EncodeStack encodes a stack of macaroons into libmacaroon v1 json format.
func (V1J) EncodeStack(stack mack.Stack) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewV1JEncoder(&buf)
	if err := enc.EncodeStack(stack); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type V1JEncoder struct {
	encoder *json.Encoder
}

func NewV1JEncoder(w io.Writer) *V1JEncoder {
	return &V1JEncoder{encoder: json.NewEncoder(w)}
}

func (enc *V1JEncoder) EncodeMacaroon(m *mack.Macaroon) error {
	js, err := v1jMacaroonToJSON(m)
	if err != nil {
		return err
	}
	return enc.encoder.Encode(js)
}

func (enc *V1JEncoder) EncodeStack(stack mack.Stack) error {
	jsonStack := make([]v1jMacaroonJSON, len(stack))
	for i := range stack {
		var err error
		jsonStack[i], err = v1jMacaroonToJSON(&stack[i])
		if err != nil {
			return err
		}
	}
	return enc.encoder.Encode(jsonStack)
}

type V1JDecoder struct {
	buf []byte
}

func NewV1JDecoder(bs []byte) *V1JDecoder {
	return &V1JDecoder{buf: bs}
}

func (dec *V1JDecoder) DecodeMacaroon(m *mack.Macaroon) error {
	var js v1jMacaroonJSON
	if err := json.Unmarshal(dec.buf, &js); err != nil {
		return fmt.Errorf("v1j.DecodeMacaroon: failed to unmarshal json: %w", err)
	}
	if err := v1jMacaroonFromJSON(&js, m); err != nil {
		return fmt.Errorf("v1j.DecodeMacaroon: failed to convert to macaroon: %w", err)
	}
	return nil
}

func (dec *V1JDecoder) DecodeStack(stack *mack.Stack) error {
	var js []v1jMacaroonJSON
	if err := json.Unmarshal(dec.buf, &js); err != nil {
		return fmt.Errorf("v1j.DecodeStack: failed to unmarshal json: %w", err)
	}
	s := make(mack.Stack, len(js))
	for i := range js {
		return v1jMacaroonFromJSON(&js[i], &s[i])
	}
	*stack = s
	return nil
}

func v1jMacaroonToJSON(m *mack.Macaroon) (v1jMacaroonJSON, error) {
	if !utf8.Valid(m.ID()) {
		return v1jMacaroonJSON{}, errors.New("v1j.EncodeMacaroon: macaroon id is not valid UTF-8")
	}
	js := v1jMacaroonJSON{
		Location:   m.Location(),
		Identifier: string(m.ID()),
		Caveats:    make([]v1jCaveatJSON, len(m.Caveats())),
		Signature:  hex.EncodeToString(m.Signature()),
	}
	for i, c := range m.Caveats() {
		cid := c.ID()
		if !utf8.Valid(cid) {
			return v1jMacaroonJSON{}, errors.New("caveat id is not valid UTF-8")
		}
		js.Caveats[i] = v1jCaveatJSON{
			Location: c.Location(),
			CID:      string(cid),
			VID:      base64.RawURLEncoding.EncodeToString(c.VID()),
		}
	}
	return js, nil
}

func v1jMacaroonFromJSON(js *v1jMacaroonJSON, m *mack.Macaroon) error {
	var raw mack.Raw
	var err error
	raw.Location = js.Location
	raw.ID = []byte(js.Identifier)
	if raw.Signature, err = hex.DecodeString(js.Signature); err != nil {
		return fmt.Errorf("v1j.DecodeMacaroon: failed to decode signature: %w", err)
	}
	raw.Caveats = make([]mack.RawCaveat, len(js.Caveats))
	for i, c := range js.Caveats {
		var vid []byte
		vid, err = Base64DecodeLoose(c.VID)
		if err != nil {
			return fmt.Errorf("v1j.DecodeMacaroon: failed to decode caveat 'vid': %w", err)
		}
		raw.Caveats[i] = mack.RawCaveat{
			CID:      []byte(c.CID),
			VID:      vid,
			Location: c.Location,
		}
	}
	*m = mack.NewFromRaw(raw)
	return nil
}

type v1jMacaroonJSON struct {
	Location   string          `json:"location"`
	Identifier string          `json:"identifier"`
	Caveats    []v1jCaveatJSON `json:"caveats"`
	Signature  string          `json:"signature"` // hex-encoded
}

// caveatJSONV1 defines the V1 JSON format for caveats within a mack.
type v1jCaveatJSON struct {
	CID      string `json:"cid"`
	VID      string `json:"vid,omitempty"`
	Location string `json:"cl,omitempty"`
}
