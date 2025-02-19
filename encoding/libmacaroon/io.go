package libmacaroon

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"io"
	"unsafe"
)

// InputDecoder decodes bytes from input data.
type InputDecoder interface {
	// DecodeInput returns a new io.Reader that decodes input data read from the given reader.
	DecodeInput(r io.Reader) io.Reader

	// DecodedLength returns the length of the decoded data based on the input length 'n'.
	DecodedLength(n int) int
}

// OutputEncoder encodes binary before writing data to the writer.
type OutputEncoder interface {
	// EncodeOutput returns a new io.Writer that encodes the data before writing it to the given writer.
	EncodeOutput(w io.Writer) io.Writer

	// EncodedLength takes an integer input 'n' and returns the encoded length.
	EncodedLength(n int) int
}

// Base64 implements OutputEncoder and InputDecoder for Base64.
type Base64 struct {
	Encoding *base64.Encoding
}

func (o *Base64) DecodeInput(r io.Reader) io.Reader {
	return base64.NewDecoder(o.Encoding, r)
}

func (o *Base64) DecodedLength(n int) int {
	return o.Encoding.DecodedLen(n)
}

func (o *Base64) EncodeOutput(w io.Writer) io.Writer {
	return base64.NewEncoder(o.Encoding, w)
}

func (o *Base64) EncodedLength(n int) int {
	return o.Encoding.EncodedLen(n)
}

// NoEncoding implements OutputEncoder and InputDecoder as no-op.
type NoEncoding struct{}

func (NoEncoding) DecodeInput(r io.Reader) io.Reader {
	return r
}

func (NoEncoding) DecodedLength(n int) int {
	return n
}

func (NoEncoding) EncodeOutput(w io.Writer) io.Writer {
	return w
}

func (NoEncoding) EncodedLength(n int) int {
	return n
}

// Hex implements OutputEncoder and InputDecoder for Hexadecimal.
type Hex struct{}

func (Hex) DecodeInput(r io.Reader) io.Reader {
	return hex.NewDecoder(r)
}

func (Hex) DecodedLength(n int) int {
	return hex.DecodedLen(n)
}

func (Hex) EncodeOutput(w io.Writer) io.Writer {
	return hex.NewEncoder(w)
}

func (Hex) EncodedLength(n int) int {
	return hex.EncodedLen(n)
}

type byteWriter struct {
	io.Writer
	buf [8]byte
}

func (w *byteWriter) WriteByte(b byte) error {
	w.buf[0] = b
	_, err := w.Write(w.buf[:1])
	return err
}

func (w *byteWriter) WriteVarint(i uint64) (int, error) {
	n := binary.PutUvarint(w.buf[:], i)
	return w.Write(w.buf[:n])
}

func (w *byteWriter) Write(b []byte) (int, error) {
	n, err := w.Writer.Write(b)
	return n, err
}

func (w *byteWriter) WriteString(str string) (int, error) {
	n, err := w.Writer.Write(unsafe.Slice(unsafe.StringData(str), len(str)))
	return n, err
}

type byteReader struct {
	buf    []byte
	offset int
}

func (br *byteReader) ReadField(n int) ([]byte, error) {
	if br.offset+n > len(br.buf) {
		return nil, io.EOF
	}
	b := br.buf[br.offset : br.offset+n]
	br.offset += n
	return b, nil
}

func (br *byteReader) ReadByte() (byte, error) {
	if br.offset >= len(br.buf) {
		return 0, io.EOF
	}
	b := br.buf[br.offset]
	br.offset++
	return b, nil
}

func (br *byteReader) Read(p []byte) (int, error) {
	if len(p) >= len(br.buf)-br.offset {
		n := copy(p, br.buf[br.offset:])
		br.offset += n
		return n, io.EOF
	}
	n := copy(p, br.buf[br.offset:])
	br.offset += n
	return n, nil
}

func decodeBuffer(dec InputDecoder, buf []byte) ([]byte, error) {
	if dec == nil {
		return buf, nil
	}
	decoded := make([]byte, dec.DecodedLength(len(buf)))
	r := dec.DecodeInput(bytes.NewReader(buf))
	_, err := io.ReadFull(r, decoded)
	return decoded, err
}

// Base64DecodeLoose tries to detect the variant and picks the appropriate base64 encoding to decode the string.
func Base64DecodeLoose(str string) ([]byte, error) {
	if str == "" {
		return nil, nil
	}
	padded := str[len(str)-1] == '='
	var url bool
	for _, b := range []byte(str) {
		if b == '-' || b == '_' {
			url = true
			break
		}
	}
	var enc *base64.Encoding
	switch {
	case padded && url:
		enc = base64.URLEncoding
	case url:
		enc = base64.RawURLEncoding
	case padded:
		enc = base64.StdEncoding
	default:
		enc = base64.RawStdEncoding
	}
	return enc.DecodeString(str)
}
