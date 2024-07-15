package macaroon

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
)

func (m *Macaroon) String() string {
	var jm jsonMacaroon
	jm.Location = jsonByteString(m.Location())
	jm.ID = m.ID()
	jm.Sig = m.Signature()
	jm.Caveat = make([]jsonCaveat, len(m.Caveats()))
	for i, c := range m.Caveats() {
		jm.Caveat[i] = jsonCaveat{
			Location: jsonByteString(c.Location()),
			VID:      c.VID(),
			CID:      c.ID(),
		}
	}
	js, _ := jsonMarshalNoEscape(jm, true)
	return string(js)
}

func (c *Caveat) String() string {
	js, _ := jsonMarshalNoEscape(jsonCaveat{
		Location: jsonByteString(c.Location()),
		VID:      c.VID(),
		CID:      c.ID(),
	}, true)
	return string(js)
}

type jsonByteString []byte

func (j jsonByteString) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte(`""`), nil
	}
	str := string(j)
	for _, r := range str {
		if r < 32 || r > 126 {
			return json.Marshal([]string{"~hex", hex.EncodeToString(j)})
		}
	}
	return jsonMarshalNoEscape(str, false)
}

type jsonMacaroon struct {
	Location jsonByteString `json:"location"`
	ID       jsonByteString `json:"id"`
	Caveat   []jsonCaveat   `json:"caveats,omitempty"`
	Sig      jsonByteString `json:"sig"`
}

type jsonCaveat struct {
	Location jsonByteString `json:"location,omitempty"`
	VID      jsonByteString `json:"vid,omitempty"`
	CID      jsonByteString `json:"cid,omitempty"`
}

// jsonMarshalNoEscape marshals a value to json without escaping HTML.
func jsonMarshalNoEscape(v any, indent bool) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if indent {
		encoder.SetIndent("", "  ")
	}
	err := encoder.Encode(v)
	return buf.Bytes(), err
}

// printableBytes returns a string representation of the bytes, convenient for debugging.
// If the size of the byteString is 0, it returns the empty string: "".
// If the string contains non-printable characters, it returns a hex-formatted string, so it can be printed.
// Otherwise, it returns the original string representation.
func printableBytes(bs []byte) string {
	if len(bs) == 0 {
		return ""
	}
	str := string(bs)
	for _, r := range str {
		if r < 32 || r > 126 {
			return "0x" + hex.EncodeToString(bs)
		}
	}
	return str
}
