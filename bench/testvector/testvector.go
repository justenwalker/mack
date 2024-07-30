package testvector

import (
	"bench/impl"
	"encoding/base64"

	"github.com/justenwalker/mack/crypt/random"
)

func RandomMacaroonSpec() impl.NewMacaroonSpec {
	return impl.NewMacaroonSpec{
		RootKey:  RandomBytes(32),
		ID:       []byte(base64.StdEncoding.EncodeToString(RandomBytes(100))),
		Location: base64.StdEncoding.EncodeToString(RandomBytes(40)),
	}
}

func SmallMacaroon() impl.NewMacaroonSpec {
	return impl.NewMacaroonSpec{
		RootKey:  RandomBytes(32),
		ID:       []byte("id:root"),
		Location: "loc://root",
		Caveats: []impl.NewCaveatSpec{
			{ID: []byte(`caveat`)},
		},
	}
}

func LargeMacaroon() impl.NewMacaroonSpec {
	return impl.NewMacaroonSpec{
		RootKey:  RandomBytes(32),
		ID:       []byte("id:flintstones"),
		Location: "The Flintstones",
		Caveats: []impl.NewCaveatSpec{
			{ID: []byte(`town=bedrock`)},
			{
				ID:       []byte("id:fred"),
				Key:      RandomBytes(32),
				Location: "Fred",
				Caveats: []impl.NewCaveatSpec{
					{ID: []byte(`stature=large`)},
					{
						ID:       []byte("id:wilma"),
						Key:      RandomBytes(32),
						Location: "Wilma",
						Caveats: []impl.NewCaveatSpec{
							{ID: []byte(`hair=red`)},
							{
								ID:       []byte("id:pebbles"),
								Key:      RandomBytes(32),
								Location: "Pebbles",
								Caveats: []impl.NewCaveatSpec{
									{ID: []byte(`hair=red`)},
								},
							},
						},
					},
				},
			},
			{
				ID:       []byte("id:barney"),
				Key:      RandomBytes(32),
				Location: "Barney",
				Caveats: []impl.NewCaveatSpec{
					{ID: []byte(`stature=short`)},
					{
						ID:       []byte("id:betty"),
						Key:      RandomBytes(32),
						Location: "Betty",
						Caveats: []impl.NewCaveatSpec{
							{ID: []byte(`hair=black`)},
							{
								ID:       []byte("id:bambam"),
								Key:      RandomBytes(32),
								Location: "Bam-Bam",
							},
						},
					},
				},
			},
		},
	}
}

func RandomBytes(n int) []byte {
	b := make([]byte, n)
	random.Read(b)
	return b
}
