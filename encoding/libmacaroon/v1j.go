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

	"github.com/justenwalker/mack/encoding"
	"github.com/justenwalker/mack/macaroon"
)

var (
	_ encoding.MacaroonEncoder = V1J{}
	_ encoding.MacaroonDecoder = V1J{}
)

type V1J struct{}

func (V1J) String() string {
	return "libmacaroon/v1j"
}

// DecodeMacaroon decodes a macaroon from libmacaroon v1 json format.
func (V1J) DecodeMacaroon(bs []byte, m *macaroon.Macaroon) error {
	br := bytes.NewReader(bs)
	dec := NewV1JDecoder(br)
	return dec.DecodeMacaroon(m)
}

// DecodeStack decodes a macaroon stack from v1 json format.
func (V1J) DecodeStack(bs []byte, stack *macaroon.Stack) error {
	br := bytes.NewReader(bs)
	dec := NewV1JDecoder(br)
	return dec.DecodeStack(stack)
}

// EncodeMacaroon encodes a macaroon into libmacaroon v1 json format.
func (V1J) EncodeMacaroon(m *macaroon.Macaroon) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewV1JEncoder(&buf)
	if err := enc.EncodeMacaroon(m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EncodeStack encodes a stack of macaroons into libmacaroon v1 json format.
func (V1J) EncodeStack(stack macaroon.Stack) ([]byte, error) {
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

func (enc *V1JEncoder) EncodeMacaroon(m *macaroon.Macaroon) error {
	js, err := v1jMacaroonToJSON(m)
	if err != nil {
		return err
	}
	return enc.encoder.Encode(js)
}

func (enc *V1JEncoder) EncodeStack(stack macaroon.Stack) error {
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
	decoder *json.Decoder
}

func NewV1JDecoder(r io.Reader) *V1JDecoder {
	return &V1JDecoder{decoder: json.NewDecoder(r)}
}

func (dec *V1JDecoder) DecodeMacaroon(m *macaroon.Macaroon) error {
	var js v1jMacaroonJSON
	if err := dec.decoder.Decode(&js); err != nil {
		return fmt.Errorf("v1j.DecodeMacaroon: failed to unmarshal json: %w", err)
	}
	if err := v1jMacaroonFromJSON(&js, m); err != nil {
		return fmt.Errorf("v1j.DecodeMacaroon: failed to convert to macaroon: %w", err)
	}
	return nil
}

func (dec *V1JDecoder) DecodeStack(stack *macaroon.Stack) error {
	var jsonstack []v1jMacaroonJSON
	err := dec.decoder.Decode(&jsonstack)
	if err != nil {
		return fmt.Errorf("v1j.DecodeStack: failed to unmarshal json: %w", err)
	}
	s := make(macaroon.Stack, len(jsonstack))
	for i := range jsonstack {
		return v1jMacaroonFromJSON(&jsonstack[i], &s[i])
	}
	*stack = s
	return nil
}

func v1jMacaroonToJSON(m *macaroon.Macaroon) (v1jMacaroonJSON, error) {
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

func v1jMacaroonFromJSON(js *v1jMacaroonJSON, m *macaroon.Macaroon) error {
	var raw macaroon.Raw
	var err error
	raw.Location = js.Location
	raw.ID = []byte(js.Identifier)
	if raw.Signature, err = hex.DecodeString(js.Signature); err != nil {
		return fmt.Errorf("v1j.DecodeMacaroon: failed to decode signature: %w", err)
	}
	raw.Caveats = make([]macaroon.RawCaveat, len(js.Caveats))
	for i, c := range js.Caveats {
		var vid []byte
		vid, err = Base64DecodeLoose(c.VID)
		if err != nil {
			return fmt.Errorf("v1j.DecodeMacaroon: failed to decode caveat 'vid': %w", err)
		}
		raw.Caveats[i] = macaroon.RawCaveat{
			CID:      []byte(c.CID),
			VID:      vid,
			Location: c.Location,
		}
	}
	*m = macaroon.NewFromRaw(raw)
	return nil
}

type v1jMacaroonJSON struct {
	Location   string          `json:"location"`
	Identifier string          `json:"identifier"`
	Caveats    []v1jCaveatJSON `json:"caveats"`
	Signature  string          `json:"signature"` // hex-encoded
}

// caveatJSONV1 defines the V1 JSON format for caveats within a macaroon.
type v1jCaveatJSON struct {
	CID      string `json:"cid"`
	VID      string `json:"vid,omitempty"`
	Location string `json:"cl,omitempty"`
}
