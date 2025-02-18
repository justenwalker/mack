package mack

import (
	"testing"
)

func TestPrintableBytes(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "nil",
			data:     nil,
			expected: "",
		},
		{
			name:     "empty",
			data:     []byte{},
			expected: "",
		},
		{
			name:     "unprintable",
			data:     []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
			expected: "0x000102030405060708090a0b0c0d0e0f",
		},
		{
			name:     "printable",
			data:     []byte(`hello`),
			expected: `hello`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := printableBytes(tt.data)
			if ps != tt.expected {
				t.Errorf("want %s, got %s", tt.expected, ps)
			}
		})
	}
}
