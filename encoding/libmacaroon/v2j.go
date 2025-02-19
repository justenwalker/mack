package libmacaroon

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode/utf8"

	"github.com/justenwalker/mack"
	"github.com/justenwalker/mack/encoding"
)

var _ encoding.EncoderDecoder = V2J{}

type V2J struct{}

func (V2J) String() string {
	return "libmacaroon/v2j"
}

// DecodeMacaroon decodes a macaroon from v2 json format.
func (V2J) DecodeMacaroon(bs []byte, m *mack.Macaroon) error {
	dec := NewV2JDecoder(bs)
	return dec.DecodeMacaroon(m)
}

// DecodeStack decodes a macaroon stack from v2 json format.
func (V2J) DecodeStack(bs []byte, stack *mack.Stack) error {
	dec := NewV2JDecoder(bs)
	return dec.DecodeStack(stack)
}

// EncodeMacaroon encodes a macaroon into libmacaroon v2 json format.
func (V2J) EncodeMacaroon(m *mack.Macaroon) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewV2JEncoder(&buf)
	if err := enc.EncodeMacaroon(m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EncodeStack encodes a stack of macaroons into libmacaroon v2 json format.
func (V2J) EncodeStack(stack mack.Stack) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewV2JEncoder(&buf)
	if err := enc.EncodeStack(stack); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type V2JEncoder struct {
	encoder *json.Encoder
}

func NewV2JEncoder(w io.Writer) *V2JEncoder {
	return &V2JEncoder{encoder: json.NewEncoder(w)}
}

func (enc *V2JEncoder) EncodeMacaroon(m *mack.Macaroon) error {
	err := enc.encoder.Encode(v2jMacaroonToJSON(m))
	if err != nil {
		return fmt.Errorf("v2j.EncodeMacaroon: failed to marshal json: %w", err)
	}
	return nil
}

func (enc *V2JEncoder) EncodeStack(stack mack.Stack) error {
	jsonStack := make([]v2jMacaroonJSON, len(stack))
	for i := range stack {
		jsonStack[i] = v2jMacaroonToJSON(&stack[i])
	}
	return enc.encoder.Encode(jsonStack)
}

type V2JDecoder struct {
	buf []byte
}

func NewV2JDecoder(bs []byte) *V2JDecoder {
	return &V2JDecoder{buf: bs}
}

func (dec *V2JDecoder) DecodeMacaroon(m *mack.Macaroon) error {
	var js v2jMacaroonJSON
	err := json.Unmarshal(dec.buf, &js)
	if err != nil {
		return fmt.Errorf("v2j.DecodeMacaroon: failed to unmarshal json: %w", err)
	}
	if js.Version != 2 {
		return fmt.Errorf("v2j.DecodeMacaroon: invalid version: %d", js.Version)
	}
	return v2jMacaroonFromJSON(&js, m)
}

func (dec *V2JDecoder) DecodeStack(stack *mack.Stack) error {
	var jsonstack []v2jMacaroonJSON
	err := json.Unmarshal(dec.buf, &jsonstack)
	if err != nil {
		return fmt.Errorf("v2j.DecodeStack: failed to unmarshal json: %w", err)
	}
	s := make(mack.Stack, len(jsonstack))
	for i := range jsonstack {
		if jsonstack[i].Version != 2 {
			return fmt.Errorf("v2j.DecodeStack: macaroon[%d]: invalid version: %d", i, jsonstack[i].Version)
		}
		if err = v2jMacaroonFromJSON(&jsonstack[i], &s[i]); err != nil {
			return fmt.Errorf("v2j.DecodeStack: macaroon[%d]: decode failed: %w", i, err)
		}
	}
	*stack = s
	return nil
}

func v2jMacaroonToJSON(m *mack.Macaroon) v2jMacaroonJSON {
	js := v2jMacaroonJSON{
		Version:  2,
		Location: m.Location(),
	}
	v2jSetData(m.ID(), &js.ID, &js.IDB64)
	v2jSetData(m.Signature(), &js.Signature, &js.SignatureB64)
	js.Caveats = make([]v2jCaveatJSON, len(m.Caveats()))
	for i, c := range m.Caveats() {
		v2jSetData(c.ID(), &js.Caveats[i].ID, &js.Caveats[i].IDB64)
		v2jSetData(c.VID(), &js.Caveats[i].Verification, &js.Caveats[i].VerificationB64)
		js.Caveats[i].Location = c.Location()
	}
	return js
}

func v2jMacaroonFromJSON(js *v2jMacaroonJSON, m *mack.Macaroon) error {
	var raw mack.Raw
	var err error
	if js.Version != 2 {
		return fmt.Errorf("v2j.DecodeMacaroon: unsupported version: %d", js.Version)
	}
	raw.ID, err = v2jJSONFieldData(js.ID, js.IDB64)
	if err != nil {
		return fmt.Errorf("v2j.DecodeMacaroon: failed to read macaroon id: %w", err)
	}
	raw.Location = js.Location
	raw.Caveats = make([]mack.RawCaveat, len(js.Caveats))
	for i, c := range js.Caveats {
		raw.Caveats[i].CID, err = v2jJSONFieldData(c.ID, c.IDB64)
		if err != nil {
			return fmt.Errorf("v2j.DecodeMacaroon: failed to read caveat[%d].cid: %w", i, err)
		}
		raw.Caveats[i].VID, err = v2jJSONFieldData(c.Verification, c.VerificationB64)
		if err != nil {
			return fmt.Errorf("v2j.DecodeMacaroon: failed to read caveat[%d].vid: %w", i, err)
		}
		raw.Caveats[i].Location = c.Location
	}
	raw.Signature, err = v2jJSONFieldData(js.Signature, js.SignatureB64)
	if err != nil {
		return fmt.Errorf("v2j.DecodeMacaroon: failed to read signature: %w", err)
	}
	*m = mack.NewFromRaw(raw)
	return nil
}

type v2jMacaroonJSON struct {
	Version      v2jVersionJSON  `json:"v"`
	Location     string          `json:"l,omitempty"`
	ID           string          `json:"i,omitempty"`
	IDB64        string          `json:"i64,omitempty"`
	Caveats      []v2jCaveatJSON `json:"c"`
	Signature    string          `json:"s,omitempty"`
	SignatureB64 string          `json:"s64,omitempty"`
}

type v2jCaveatJSON struct {
	ID              string `json:"i,omitempty"`
	IDB64           string `json:"i64,omitempty"`
	Location        string `json:"l,omitempty"`
	Verification    string `json:"v,omitempty"`
	VerificationB64 string `json:"v64,omitempty"`
}

type v2jVersionJSON int

func (j *v2jVersionJSON) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	if b[0] == '"' {
		var str string
		if err := json.Unmarshal(b, &str); err != nil {
			return err
		}
		v, err := strconv.Atoi(str)
		if err != nil {
			return fmt.Errorf("v2j.DecodeMacaroon: failed to unmarshal version: %w", err)
		}
		*j = v2jVersionJSON(v)
		return nil
	}
	var v int
	if err := json.Unmarshal(b, &v); err != nil {
		return fmt.Errorf("v2j.DecodeMacaroon: failed to unmarshal version: %w", err)
	}
	*j = v2jVersionJSON(v)
	return nil
}

func v2jSetData(data []byte, sp *string, b64p *string) {
	if len(data) == 0 {
		return
	}
	if utf8.Valid(data) {
		*sp = string(data)
		return
	}
	*b64p = base64.RawURLEncoding.EncodeToString(data)
}

func v2jJSONFieldData(str string, b64 string) ([]byte, error) {
	if str != "" && b64 != "" {
		return nil, errors.New("x and x64 fields are mutually exclusive")
	}
	if str == "" && b64 == "" {
		return nil, nil
	}
	if str != "" {
		return []byte(str), nil
	}
	return Base64DecodeLoose(b64)
}
