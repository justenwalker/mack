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

	"github.com/justenwalker/mack/encoding"
	"github.com/justenwalker/mack/macaroon"
)

var (
	_ encoding.MacaroonEncoder = V2J{}
	_ encoding.MacaroonDecoder = V2J{}
)

type V2J struct{}

func (V2J) String() string {
	return "libmacaroon/v2j"
}

// DecodeMacaroon decodes a macaroon from v2 json format.
func (V2J) DecodeMacaroon(bs []byte, m *macaroon.Macaroon) error {
	br := bytes.NewReader(bs)
	dec := NewV2JDecoder(br)
	return dec.DecodeMacaroon(m)
}

// DecodeStack decodes a macaroon stack from v2 json format.
func (V2J) DecodeStack(bs []byte, stack *macaroon.Stack) error {
	br := bytes.NewReader(bs)
	dec := NewV2JDecoder(br)
	return dec.DecodeStack(stack)
}

// EncodeMacaroon encodes a macaroon into libmacaroon v2 json format.
func (V2J) EncodeMacaroon(m *macaroon.Macaroon) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewV2JEncoder(&buf)
	if err := enc.EncodeMacaroon(m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EncodeStack encodes a stack of macaroons into libmacaroon v2 json format.
func (V2J) EncodeStack(stack macaroon.Stack) ([]byte, error) {
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

func (enc *V2JEncoder) EncodeMacaroon(m *macaroon.Macaroon) error {
	err := enc.encoder.Encode(v2jMacaroonToJSON(m))
	if err != nil {
		return fmt.Errorf("v2j.EncodeMacaroon: failed to marshal json: %w", err)
	}
	return nil
}

func (enc *V2JEncoder) EncodeStack(stack macaroon.Stack) error {
	jsonStack := make([]v2jMacaroonJSON, len(stack))
	for i := range stack {
		jsonStack[i] = v2jMacaroonToJSON(&stack[i])
	}
	return enc.encoder.Encode(jsonStack)
}

type V2JDecoder struct {
	decoder *json.Decoder
}

func NewV2JDecoder(r io.Reader) *V2JDecoder {
	return &V2JDecoder{decoder: json.NewDecoder(r)}
}

func (dec *V2JDecoder) DecodeMacaroon(m *macaroon.Macaroon) error {
	var js v2jMacaroonJSON
	err := dec.decoder.Decode(&js)
	if err != nil {
		return fmt.Errorf("v2j.DecodeMacaroon: failed to unmarshal json: %w", err)
	}
	return v2jMacaroonFromJSON(&js, m)
}

func (dec *V2JDecoder) DecodeStack(stack *macaroon.Stack) error {
	var jsonstack []v2jMacaroonJSON
	err := dec.decoder.Decode(&jsonstack)
	if err != nil {
		return fmt.Errorf("v2j.DecodeStack: failed to unmarshal json: %w", err)
	}
	s := make(macaroon.Stack, len(jsonstack))
	for i := range jsonstack {
		return v2jMacaroonFromJSON(&jsonstack[i], &s[i])
	}
	*stack = s
	return nil
}

func v2jMacaroonToJSON(m *macaroon.Macaroon) v2jMacaroonJSON {
	id, id64 := v2jJSONData(m.ID())
	sig, sig64 := v2jJSONData(m.Signature())
	cs := make([]v2jCaveatJSON, len(m.Caveats()))
	for i, c := range m.Caveats() {
		cid, cid64 := v2jJSONData(c.ID())
		vid, vid64 := v2jJSONData(c.VID())
		cs[i] = v2jCaveatJSON{
			ID:              cid,
			IDB64:           cid64,
			Location:        c.Location(),
			Verification:    vid,
			VerificationB64: vid64,
		}
	}
	return v2jMacaroonJSON{
		Version:      2,
		ID:           id,
		IDB64:        id64,
		Location:     m.Location(),
		Caveats:      cs,
		Signature:    sig,
		SignatureB64: sig64,
	}
}

func v2jMacaroonFromJSON(js *v2jMacaroonJSON, m *macaroon.Macaroon) error {
	var raw macaroon.Raw
	var err error
	if js.Version != 2 {
		return fmt.Errorf("v2j.DecodeMacaroon: unsupported version: %d", js.Version)
	}
	raw.ID, err = v2jJSONFieldData(js.ID, js.IDB64)
	if err != nil {
		return fmt.Errorf("v2j.DecodeMacaroon: failed to read macaroon id: %w", err)
	}
	raw.Location = js.Location
	raw.Caveats = make([]macaroon.RawCaveat, len(js.Caveats))
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
	*m = macaroon.NewFromRaw(raw)
	return nil
}

type v2jMacaroonJSON struct {
	Version      v2jVersionJSON  `json:"v"`
	Location     string          `json:"l,omitempty"`
	ID           *string         `json:"i,omitempty"`
	IDB64        *string         `json:"i64,omitempty"`
	Caveats      []v2jCaveatJSON `json:"c"`
	Signature    *string         `json:"s,omitempty"`
	SignatureB64 *string         `json:"s64,omitempty"`
}

type v2jCaveatJSON struct {
	ID              *string `json:"i,omitempty"`
	IDB64           *string `json:"i64,omitempty"`
	Location        string  `json:"l,omitempty"`
	Verification    *string `json:"v,omitempty"`
	VerificationB64 *string `json:"v64,omitempty"`
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

func v2jJSONData(data []byte) (str *string, b64 *string) {
	if len(data) == 0 {
		return nil, nil
	}
	if utf8.Valid(data) {
		s := string(data)
		return &s, nil
	}
	b := base64.RawURLEncoding.EncodeToString(data)
	return nil, &b
}

func v2jJSONFieldData(str *string, b64 *string) ([]byte, error) {
	if str != nil && b64 != nil {
		return nil, errors.New("x and x64 fields are mutually exclusive")
	}
	if str == nil && b64 == nil {
		return nil, nil
	}
	if str != nil {
		return []byte(*str), nil
	}
	return Base64DecodeLoose(*b64)
}
