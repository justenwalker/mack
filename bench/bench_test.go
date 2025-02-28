package bench

import (
	"encoding/base64"
	"testing"

	"bench/impl"
	"bench/impl/libmacaroon"
	"bench/impl/mack"
	"bench/testvector"
)

func createImplementations(tb testing.TB) []impl.Implementation {
	tb.Helper()
	impls := []impl.Implementation{
		{
			Name:      "libmacaroon",
			Interface: &libmacaroon.Implementation{},
		},
		{
			Name:      "mack",
			Interface: &mack.Implementation{},
		},
	}
	for i := range impls {
		if err := impls[i].Setup(); err != nil {
			tb.Fatalf("impl.Setup(%s): %v", impls[i].Name, err)
		}
	}
	return impls
}

func BenchmarkNewMacaroon(b *testing.B) {
	impls := createImplementations(b)
	args := testvector.RandomMacaroonSpec()
	for _, im := range impls {
		b.Run("impl="+im.Name, func(b *testing.B) {
			b.ReportAllocs()
			var err error
			for b.Loop() {
				_, err = im.NewMacaroons(args)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkAddFirstPartyCaveat(b *testing.B) {
	impls := createImplementations(b)
	args := testvector.RandomMacaroonSpec()
	cid := []byte(base64.StdEncoding.EncodeToString(testvector.RandomBytes(100)))
	for _, im := range impls {
		b.Run("impl="+im.Name, func(b *testing.B) {
			m, err := im.NewMacaroon(args)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			for b.Loop() {
				_, err = im.AddFirstPartyCaveat(m, cid)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkVerify_small(b *testing.B) {
	impls := createImplementations(b)
	args := testvector.SmallMacaroon()
	for _, im := range impls {
		b.Run("impl="+im.Name, func(b *testing.B) {
			m, err := im.NewMacaroons(args)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			for b.Loop() {
				if _, err = im.VerifyMacaroon(args.RootKey, m); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkVerify_large(b *testing.B) {
	impls := createImplementations(b)
	args := testvector.LargeMacaroon()
	for _, im := range impls {
		b.Run("impl="+im.Name, func(b *testing.B) {
			m, err := im.NewMacaroons(args)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			for b.Loop() {
				if _, err = im.VerifyMacaroon(args.RootKey, m); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkEncodeToV2J(b *testing.B) {
	impls := createImplementations(b)
	args := testvector.RandomMacaroonSpec()
	for _, im := range impls {
		b.Run("impl="+im.Name, func(b *testing.B) {
			m, err := im.NewMacaroon(args)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			for b.Loop() {
				if _, err = im.EncodeToV2J(m); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkEncodeToV2(b *testing.B) {
	impls := createImplementations(b)
	args := testvector.RandomMacaroonSpec()
	for _, im := range impls {
		b.Run("impl="+im.Name, func(b *testing.B) {
			m, err := im.NewMacaroon(args)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			for b.Loop() {
				if _, err = im.EncodeToV2(m); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDecodeFromV2J(b *testing.B) {
	impls := createImplementations(b)
	args := testvector.RandomMacaroonSpec()
	for _, im := range impls {
		b.Run("impl="+im.Name, func(b *testing.B) {
			m, err := im.NewMacaroon(args)
			if err != nil {
				b.Fatal(err)
			}
			encoded, err := im.EncodeToV2J(m)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			for b.Loop() {
				if _, err = im.DecodeFromV2J(encoded); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDecodeFromV2(b *testing.B) {
	impls := createImplementations(b)
	args := testvector.RandomMacaroonSpec()
	for _, im := range impls {
		b.Run("impl="+im.Name, func(b *testing.B) {
			m, err := im.NewMacaroon(args)
			if err != nil {
				b.Fatal(err)
			}
			encoded, err := im.EncodeToV2(m)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			for b.Loop() {
				if _, err = im.DecodeFromV2(encoded); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
