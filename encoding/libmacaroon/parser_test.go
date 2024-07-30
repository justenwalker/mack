package libmacaroon_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/justenwalker/mack/encoding"
	"github.com/justenwalker/mack/encoding/libmacaroon"
	"github.com/justenwalker/mack/macaroon"
	"github.com/justenwalker/mack/sensible"
)

func TestParser_DecodeMacaroon(t *testing.T) {
	type parseTest struct {
		name string
		b64  string
	}
	p := libmacaroon.Parser{}
	tests := []struct {
		name      string
		encodings []parseTest
	}{
		{
			name: "serialization_1",
			encodings: []parseTest{
				{name: "v1", b64: "TURBeU1XeHZZMkYwYVc5dUlHaDBkSEE2THk5bGVHRnRjR3hsTG05eVp5OEtNREF4Tldsa1pXNTBhV1pwWlhJZ2EyVjVhV1FLTURBeVpuTnBaMjVoZEhWeVpTQjgzdWVTVVJ4Ynh2VW9TRmdGMy1teVRuaGVLT0twa3dINTF4SEdDZU9POXdv"},
				{name: "v2", b64: "AgETaHR0cDovL2V4YW1wbGUub3JnLwIFa2V5aWQAAAYgfN7nklEcW8b1KEhYBd_psk54XijiqZMB-dcRxgnjjvc="},
				{name: "v2j", b64: "eyJ2IjoyLCJsIjoiaHR0cDovL2V4YW1wbGUub3JnLyIsImkiOiJrZXlpZCIsImMiOltdLCJzNjQiOiJmTjdua2xFY1c4YjFLRWhZQmRfcHNrNTRYaWppcVpNQi1kY1J4Z25qanZjIn0"},
			},
		},
		{
			name: "serialization_2",
			encodings: []parseTest{
				{name: "v1", b64: "TURBeU1XeHZZMkYwYVc5dUlHaDBkSEE2THk5bGVHRnRjR3hsTG05eVp5OEtNREF4Tldsa1pXNTBhV1pwWlhJZ2EyVjVhV1FLTURBeFpHTnBaQ0JoWTJOdmRXNTBJRDBnTXpjek5Ua3lPRFUxT1Fvd01ESm1jMmxuYm1GMGRYSmxJUFZJQl9iY2J0LUl2dzl6QnJPQ0pXS2pZbE05djNNNXVtRjJYYVM5SloySENn"},
				{name: "v2", b64: "AgETaHR0cDovL2V4YW1wbGUub3JnLwIFa2V5aWQAAhRhY2NvdW50ID0gMzczNTkyODU1OQAABiD1SAf23G7fiL8PcwazgiVio2JTPb9zObphdl2kvSWdhw=="},
				{name: "v2j", b64: "eyJ2IjoyLCJsIjoiaHR0cDovL2V4YW1wbGUub3JnLyIsImkiOiJrZXlpZCIsImMiOlt7ImkiOiJhY2NvdW50ID0gMzczNTkyODU1OSJ9XSwiczY0IjoiOVVnSDl0eHUzNGlfRDNNR3M0SWxZcU5pVXoyX2N6bTZZWFpkcEwwbG5ZYyJ9"},
			},
		},
		{
			name: "serialization_3",
			encodings: []parseTest{
				{name: "v1", b64: "TURBeU1XeHZZMkYwYVc5dUlHaDBkSEE2THk5bGVHRnRjR3hsTG05eVp5OEtNREF4Tldsa1pXNTBhV1pwWlhJZ2EyVjVhV1FLTURBeFpHTnBaQ0JoWTJOdmRXNTBJRDBnTXpjek5Ua3lPRFUxT1Fvd01ERTFZMmxrSUhWelpYSWdQU0JoYkdsalpRb3dNREptYzJsbmJtRjBkWEpsSUV2cFo4MGVvTWF5YTY5cVNwVHVtd1d4V0liYUM2aGVqRUtwUEkwT0VsNzhDZw=="},
				{name: "v2", b64: "AgETaHR0cDovL2V4YW1wbGUub3JnLwIFa2V5aWQAAhRhY2NvdW50ID0gMzczNTkyODU1OQACDHVzZXIgPSBhbGljZQAABiBL6WfNHqDGsmuvakqU7psFsViG2guoXoxCqTyNDhJe_A=="},
				{name: "v2j", b64: "eyJ2IjoyLCJsIjoiaHR0cDovL2V4YW1wbGUub3JnLyIsImkiOiJrZXlpZCIsImMiOlt7ImkiOiJhY2NvdW50ID0gMzczNTkyODU1OSJ9LHsiaSI6InVzZXIgPSBhbGljZSJ9XSwiczY0IjoiUy1sbnpSNmd4ckpycjJwS2xPNmJCYkZZaHRvTHFGNk1RcWs4alE0U1h2dyJ9"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			macaroons := make([]macaroon.Macaroon, len(tt.encodings))
			serialized := make([][]byte, len(tt.encodings))
			for i := range tt.encodings {
				t.Run(tt.encodings[i].name, func(t *testing.T) {
					var err error
					serialized[i], err = libmacaroon.Base64DecodeLoose(tt.encodings[i].b64)
					if err != nil {
						t.Fatalf("encoding[%d]: base64 decoding failed: %v", i, err)
					}
					if err = p.DecodeMacaroon(serialized[i], &macaroons[i]); err != nil {
						t.Fatalf("encoding[%d]: failed to decode: %v", i, err)
					}
				})
			}
		})
	}
}

func TestSerialization(t *testing.T) {
	type encoderDecoder interface {
		encoding.MacaroonEncoder
		encoding.MacaroonDecoder
	}
	type serializationTest struct {
		encoding  encoderDecoder
		b64       string
		normalize func([]byte) string
	}
	normalizeTrimNewline := func(bs []byte) string {
		return strings.TrimSuffix(string(bs), "\n")
	}
	normalizeAsIs := func(bs []byte) string {
		return string(bs)
	}

	b64 := &libmacaroon.Base64{Encoding: base64.RawURLEncoding}
	v1 := libmacaroon.V1{
		OutputEncoder: b64,
		InputDecoder:  b64,
	}
	v2 := libmacaroon.V2{}
	v2j := libmacaroon.V2J{}
	tests := []struct {
		name      string
		encodings []serializationTest
	}{
		{
			name: "serialization_1",
			encodings: []serializationTest{
				{encoding: v1, b64: "TURBeU1XeHZZMkYwYVc5dUlHaDBkSEE2THk5bGVHRnRjR3hsTG05eVp5OEtNREF4Tldsa1pXNTBhV1pwWlhJZ2EyVjVhV1FLTURBeVpuTnBaMjVoZEhWeVpTQjgzdWVTVVJ4Ynh2VW9TRmdGMy1teVRuaGVLT0twa3dINTF4SEdDZU9POXdv", normalize: normalizeAsIs},
				{encoding: v2, b64: "AgETaHR0cDovL2V4YW1wbGUub3JnLwIFa2V5aWQAAAYgfN7nklEcW8b1KEhYBd_psk54XijiqZMB-dcRxgnjjvc=", normalize: normalizeAsIs},
				{encoding: v2j, b64: "eyJ2IjoyLCJsIjoiaHR0cDovL2V4YW1wbGUub3JnLyIsImkiOiJrZXlpZCIsImMiOltdLCJzNjQiOiJmTjdua2xFY1c4YjFLRWhZQmRfcHNrNTRYaWppcVpNQi1kY1J4Z25qanZjIn0", normalize: normalizeTrimNewline},
			},
		},
		{
			name: "serialization_2",
			encodings: []serializationTest{
				{encoding: v1, b64: "TURBeU1XeHZZMkYwYVc5dUlHaDBkSEE2THk5bGVHRnRjR3hsTG05eVp5OEtNREF4Tldsa1pXNTBhV1pwWlhJZ2EyVjVhV1FLTURBeFpHTnBaQ0JoWTJOdmRXNTBJRDBnTXpjek5Ua3lPRFUxT1Fvd01ESm1jMmxuYm1GMGRYSmxJUFZJQl9iY2J0LUl2dzl6QnJPQ0pXS2pZbE05djNNNXVtRjJYYVM5SloySENn", normalize: normalizeAsIs},
				{encoding: v2, b64: "AgETaHR0cDovL2V4YW1wbGUub3JnLwIFa2V5aWQAAhRhY2NvdW50ID0gMzczNTkyODU1OQAABiD1SAf23G7fiL8PcwazgiVio2JTPb9zObphdl2kvSWdhw==", normalize: normalizeAsIs},
				{encoding: v2j, b64: "eyJ2IjoyLCJsIjoiaHR0cDovL2V4YW1wbGUub3JnLyIsImkiOiJrZXlpZCIsImMiOlt7ImkiOiJhY2NvdW50ID0gMzczNTkyODU1OSJ9XSwiczY0IjoiOVVnSDl0eHUzNGlfRDNNR3M0SWxZcU5pVXoyX2N6bTZZWFpkcEwwbG5ZYyJ9", normalize: normalizeTrimNewline},
			},
		},
		{
			name: "serialization_3",
			encodings: []serializationTest{
				{encoding: v1, b64: "TURBeU1XeHZZMkYwYVc5dUlHaDBkSEE2THk5bGVHRnRjR3hsTG05eVp5OEtNREF4Tldsa1pXNTBhV1pwWlhJZ2EyVjVhV1FLTURBeFpHTnBaQ0JoWTJOdmRXNTBJRDBnTXpjek5Ua3lPRFUxT1Fvd01ERTFZMmxrSUhWelpYSWdQU0JoYkdsalpRb3dNREptYzJsbmJtRjBkWEpsSUV2cFo4MGVvTWF5YTY5cVNwVHVtd1d4V0liYUM2aGVqRUtwUEkwT0VsNzhDZw==", normalize: normalizeAsIs},
				{encoding: v2, b64: "AgETaHR0cDovL2V4YW1wbGUub3JnLwIFa2V5aWQAAhRhY2NvdW50ID0gMzczNTkyODU1OQACDHVzZXIgPSBhbGljZQAABiBL6WfNHqDGsmuvakqU7psFsViG2guoXoxCqTyNDhJe_A==", normalize: normalizeAsIs},
				{encoding: v2j, b64: "eyJ2IjoyLCJsIjoiaHR0cDovL2V4YW1wbGUub3JnLyIsImkiOiJrZXlpZCIsImMiOlt7ImkiOiJhY2NvdW50ID0gMzczNTkyODU1OSJ9LHsiaSI6InVzZXIgPSBhbGljZSJ9XSwiczY0IjoiUy1sbnpSNmd4ckpycjJwS2xPNmJCYkZZaHRvTHFGNk1RcWs4alE0U1h2dyJ9", normalize: normalizeTrimNewline},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			macaroons := make([]macaroon.Macaroon, len(tt.encodings))
			serialized := make([][]byte, len(tt.encodings))
			for i := range macaroons {
				var err error
				serialized[i], err = libmacaroon.Base64DecodeLoose(tt.encodings[i].b64)
				if err != nil {
					t.Fatalf("encoding[%d]: base64 decoding failed: %v", i, err)
				}
				if err = tt.encodings[i].encoding.DecodeMacaroon(serialized[i], &macaroons[i]); err != nil {
					t.Fatalf("encoding[%d]: failed to decode: %v", i, err)
				}
			}
			for i := range macaroons {
				j := (i + 1) % len(macaroons)
				if !macaroons[i].Equal(&macaroons[j]) {
					t.Errorf("macaroon[%d] != macaroon[%d]:\n%s", i, j, cmp.Diff((&macaroons[i]).String(), (&macaroons[j]).String()))
				}
				t.Logf("macaroon[%d] == macaroon[%d]", i, j)
				encodedBytes, err := tt.encodings[i].encoding.EncodeMacaroon(&macaroons[i])
				if err != nil {
					t.Fatalf("macaroon[%d]: failed to encode: %v", i, err)
				}
				encodedString := tt.encodings[i].normalize(encodedBytes)
				if diff := cmp.Diff(string(serialized[i]), encodedString); diff != "" {
					t.Fatalf("macaroon[%d]: encoding '%v' did not match: %s", i, tt.encodings[i].encoding, diff)
				}
				t.Log(macaroons[i].String())
			}
		})
	}
}

func TestLibmacaroon(t *testing.T) {
	b64 := &libmacaroon.Base64{Encoding: base64.RawURLEncoding}
	sch := sensible.Scheme()
	type encoderDecoder interface {
		encoding.MacaroonEncoder
		encoding.MacaroonDecoder
	}
	var key [sha256.Size]byte
	copy(key[:], "macaroons-key-generator")
	keygen := hmac.New(sha256.New, key[:])
	v1 := libmacaroon.V1{
		OutputEncoder: b64,
		InputDecoder:  b64,
	}
	v2 := libmacaroon.V2{}
	tests := []struct {
		name       string
		authorized bool
		key        string
		caveats    []macaroon.PredicateChecker
		base64data string
		encoding   encoderDecoder
	}{
		{
			name:       "caveat_v1_1",
			authorized: true,
			key:        "this is the key",
			caveats: []macaroon.PredicateChecker{
				Exact("account = 3735928559"),
			},
			encoding:   v1,
			base64data: "TURBeU1XeHZZMkYwYVc5dUlHaDBkSEE2THk5bGVHRnRjR3hsTG05eVp5OEtNREF4Tldsa1pXNTBhV1pwWlhJZ2EyVjVhV1FLTURBeFpHTnBaQ0JoWTJOdmRXNTBJRDBnTXpjek5Ua3lPRFUxT1Fvd01ESm1jMmxuYm1GMGRYSmxJUFZJQl9iY2J0LUl2dzl6QnJPQ0pXS2pZbE05djNNNXVtRjJYYVM5SloySENn",
		},
		{
			name:       "caveat_v1_2",
			authorized: false,
			key:        "this is the key",
			caveats: []macaroon.PredicateChecker{
				Exact("account = 0000000000"),
			},
			encoding:   v1,
			base64data: "TURBeU1XeHZZMkYwYVc5dUlHaDBkSEE2THk5bGVHRnRjR3hsTG05eVp5OEtNREF4Tldsa1pXNTBhV1pwWlhJZ2EyVjVhV1FLTURBeFpHTnBaQ0JoWTJOdmRXNTBJRDBnTXpjek5Ua3lPRFUxT1Fvd01ESm1jMmxuYm1GMGRYSmxJUFZJQl9iY2J0LUl2dzl6QnJPQ0pXS2pZbE05djNNNXVtRjJYYVM5SloySENn",
		},
		{
			name:       "caveat_v1_3",
			authorized: false,
			key:        "this is the key",
			encoding:   v1,
			base64data: "TURBeU1XeHZZMkYwYVc5dUlHaDBkSEE2THk5bGVHRnRjR3hsTG05eVp5OEtNREF4Tldsa1pXNTBhV1pwWlhJZ2EyVjVhV1FLTURBeFpHTnBaQ0JoWTJOdmRXNTBJRDBnTXpjek5Ua3lPRFUxT1Fvd01ESm1jMmxuYm1GMGRYSmxJUFZJQl9iY2J0LUl2dzl6QnJPQ0pXS2pZbE05djNNNXVtRjJYYVM5SloySENn",
		},
		{
			name:       "caveat_v1_4",
			authorized: true,
			key:        "this is the key",
			encoding:   v1,
			caveats: []macaroon.PredicateChecker{
				Exact("account = 3735928559"),
				Exact("user = alice"),
			},
			base64data: "TURBeU1XeHZZMkYwYVc5dUlHaDBkSEE2THk5bGVHRnRjR3hsTG05eVp5OEtNREF4Tldsa1pXNTBhV1pwWlhJZ2EyVjVhV1FLTURBeFpHTnBaQ0JoWTJOdmRXNTBJRDBnTXpjek5Ua3lPRFUxT1Fvd01ERTFZMmxrSUhWelpYSWdQU0JoYkdsalpRb3dNREptYzJsbmJtRjBkWEpsSUV2cFo4MGVvTWF5YTY5cVNwVHVtd1d4V0liYUM2aGVqRUtwUEkwT0VsNzhDZw==",
		},
		{
			name:       "caveat_v1_5",
			authorized: false,
			key:        "this is the key",
			encoding:   v1,
			caveats: []macaroon.PredicateChecker{
				Exact("account = 3735928559"),
			},
			base64data: "TURBeU1XeHZZMkYwYVc5dUlHaDBkSEE2THk5bGVHRnRjR3hsTG05eVp5OEtNREF4Tldsa1pXNTBhV1pwWlhJZ2EyVjVhV1FLTURBeFpHTnBaQ0JoWTJOdmRXNTBJRDBnTXpjek5Ua3lPRFUxT1Fvd01ERTFZMmxrSUhWelpYSWdQU0JoYkdsalpRb3dNREptYzJsbmJtRjBkWEpsSUV2cFo4MGVvTWF5YTY5cVNwVHVtd1d4V0liYUM2aGVqRUtwUEkwT0VsNzhDZw==",
		},
		{
			name:       "caveat_v1_6",
			authorized: false,
			key:        "this is the key",
			encoding:   v1,
			caveats: []macaroon.PredicateChecker{
				Exact("user = alice"),
			},
			base64data: "TURBeU1XeHZZMkYwYVc5dUlHaDBkSEE2THk5bGVHRnRjR3hsTG05eVp5OEtNREF4Tldsa1pXNTBhV1pwWlhJZ2EyVjVhV1FLTURBeFpHTnBaQ0JoWTJOdmRXNTBJRDBnTXpjek5Ua3lPRFUxT1Fvd01ERTFZMmxrSUhWelpYSWdQU0JoYkdsalpRb3dNREptYzJsbmJtRjBkWEpsSUV2cFo4MGVvTWF5YTY5cVNwVHVtd1d4V0liYUM2aGVqRUtwUEkwT0VsNzhDZw==",
		},
		{
			name:       "caveat_v2_1",
			authorized: true,
			key:        "this is the key",
			caveats: []macaroon.PredicateChecker{
				Exact("account = 3735928559"),
			},
			encoding:   v2,
			base64data: "AgETaHR0cDovL2V4YW1wbGUub3JnLwIFa2V5aWQAAhRhY2NvdW50ID0gMzczNTkyODU1OQAABiD1SAf23G7fiL8PcwazgiVio2JTPb9zObphdl2kvSWdhw",
		},
		{
			name:       "caveat_v2_2",
			authorized: false,
			key:        "this is the key",
			caveats: []macaroon.PredicateChecker{
				Exact("account = 0000000000"),
			},
			encoding:   v2,
			base64data: "AgETaHR0cDovL2V4YW1wbGUub3JnLwIFa2V5aWQAAhRhY2NvdW50ID0gMzczNTkyODU1OQAABiD1SAf23G7fiL8PcwazgiVio2JTPb9zObphdl2kvSWdhw",
		},
		{
			name:       "caveat_v2_3",
			authorized: false,
			key:        "this is the key",
			encoding:   v2,
			base64data: "AgETaHR0cDovL2V4YW1wbGUub3JnLwIFa2V5aWQAAhRhY2NvdW50ID0gMzczNTkyODU1OQAABiD1SAf23G7fiL8PcwazgiVio2JTPb9zObphdl2kvSWdhw",
		},
		{
			name:       "caveat_v2_4",
			authorized: true,
			key:        "this is the key",
			encoding:   v2,
			caveats: []macaroon.PredicateChecker{
				Exact("account = 3735928559"),
				Exact("user = alice"),
			},
			base64data: "AgETaHR0cDovL2V4YW1wbGUub3JnLwIFa2V5aWQAAhRhY2NvdW50ID0gMzczNTkyODU1OQACDHVzZXIgPSBhbGljZQAABiBL6WfNHqDGsmuvakqU7psFsViG2guoXoxCqTyNDhJe_A",
		},
		{
			name:       "caveat_v2_5",
			authorized: false,
			key:        "this is the key",
			encoding:   v2,
			caveats: []macaroon.PredicateChecker{
				Exact("account = 3735928559"),
			},
			base64data: "AgETaHR0cDovL2V4YW1wbGUub3JnLwIFa2V5aWQAAhRhY2NvdW50ID0gMzczNTkyODU1OQACDHVzZXIgPSBhbGljZQAABiBL6WfNHqDGsmuvakqU7psFsViG2guoXoxCqTyNDhJe_A",
		},
		{
			name:       "caveat_v2_6",
			authorized: false,
			key:        "this is the key",
			encoding:   v2,
			caveats: []macaroon.PredicateChecker{
				Exact("user = alice"),
			},
			base64data: "AgETaHR0cDovL2V4YW1wbGUub3JnLwIFa2V5aWQAAhRhY2NvdW50ID0gMzczNTkyODU1OQACDHVzZXIgPSBhbGljZQAABiBL6WfNHqDGsmuvakqU7psFsViG2guoXoxCqTyNDhJe_A",
		},
		{
			name:       "root_v1_1",
			authorized: true,
			key:        "this is the key",
			encoding:   v1,
			base64data: "TURBeU1XeHZZMkYwYVc5dUlHaDBkSEE2THk5bGVHRnRjR3hsTG05eVp5OEtNREF4Tldsa1pXNTBhV1pwWlhJZ2EyVjVhV1FLTURBeVpuTnBaMjVoZEhWeVpTQjgzdWVTVVJ4Ynh2VW9TRmdGMy1teVRuaGVLT0twa3dINTF4SEdDZU9POXdv",
		},
		{
			name:       "root_v1_2",
			authorized: false,
			key:        "this is not the key",
			encoding:   v1,
			base64data: "TURBeU1XeHZZMkYwYVc5dUlHaDBkSEE2THk5bGVHRnRjR3hsTG05eVp5OEtNREF4Tldsa1pXNTBhV1pwWlhJZ2EyVjVhV1FLTURBeVpuTnBaMjVoZEhWeVpTQjgzdWVTVVJ4Ynh2VW9TRmdGMy1teVRuaGVLT0twa3dINTF4SEdDZU9POXdv",
		},
		{
			name:       "root_v2_1",
			authorized: true,
			key:        "this is the key",
			encoding:   v2,
			base64data: "AgETaHR0cDovL2V4YW1wbGUub3JnLwIFa2V5aWQAAAYgfN7nklEcW8b1KEhYBd_psk54XijiqZMB-dcRxgnjjvc",
		},
		{
			name:       "root_v2_2",
			authorized: false,
			key:        "this is not the key",
			encoding:   v2,
			base64data: "AgETaHR0cDovL2V4YW1wbGUub3JnLwIFa2V5aWQAAAYgfN7nklEcW8b1KEhYBd_psk54XijiqZMB-dcRxgnjjvc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := libmacaroon.Base64DecodeLoose(tt.base64data)
			if err != nil {
				t.Fatalf("base64 decoding failed: %v", err)
			}
			keygen.Reset()
			keygen.Write([]byte(tt.key))
			rootKey := keygen.Sum(nil)
			var m macaroon.Macaroon
			if err = tt.encoding.DecodeMacaroon(data, &m); err != nil {
				t.Fatalf("macaroon decoding failed: %v", err)
			}
			ctx := macaroon.WithVerifyContext(context.Background())
			stack, _ := sch.PrepareStack(&m, nil)
			vc, err := sch.Verify(ctx, rootKey, stack)
			if err != nil {
				if tt.authorized {
					traces := macaroon.GetTraces(ctx)
					t.Fatalf("expected authorization to succeed, but failed: %v\n%s", err, traces.String())
				}
				return // unauthorized
			}
			if err = vc.Clear(ctx, PredicateChecker(tt.caveats)); err != nil {
				if tt.authorized {
					t.Fatalf("expected authorization to succeed, but failed: %v", err)
				}
				return // unauthorized
			}
			if !tt.authorized {
				t.Fatalf("expected authorization to fail, but succeeded")
			}
		})
	}
}

type Exact string

func (e Exact) CheckPredicate(_ context.Context, predicate []byte) (bool, error) {
	return bytes.Equal([]byte(e), predicate), nil
}

type PredicateChecker []macaroon.PredicateChecker

func (e PredicateChecker) CheckPredicate(ctx context.Context, predicate []byte) (bool, error) {
	for _, pc := range e {
		if ok, _ := pc.CheckPredicate(ctx, predicate); ok {
			return true, nil
		}
	}
	return false, nil
}

var _ macaroon.PredicateChecker = Exact("")
