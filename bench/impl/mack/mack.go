package mack

import (
	"context"
	"fmt"

	"bench/impl"

	"github.com/justenwalker/mack/encoding/libmacaroon"
	"github.com/justenwalker/mack/macaroon"
	"github.com/justenwalker/mack/sensible"
)

var _ impl.Interface = (*Implementation)(nil)

var mackScheme = sensible.Scheme()

type Implementation struct{}

func (l *Implementation) Setup() error {
	return nil
}

func (l *Implementation) NewMacaroon(args impl.NewMacaroonSpec) (impl.Macaroon, error) {
	m, err := mackScheme.UnsafeRootMacaroon(args.Location, args.ID, args.RootKey)
	if err != nil {
		return impl.Macaroon{}, fmt.Errorf("mack/macaroon.UnsafeRootMacaroon: %w", err)
	}
	for _, c := range args.Caveats {
		m, err = mackScheme.AddFirstPartyCaveat(&m, c.ID)
		if err != nil {
			return impl.Macaroon{}, fmt.Errorf("mack/macaroon.AddFirstPartyCaveat: %w", err)
		}
	}
	return impl.Macaroon{Macaroon: &m}, nil
}

func (l *Implementation) NewMacaroons(args impl.NewMacaroonSpec) (impl.Macaroons, error) {
	ms, err := l.newMacaroonsFromSpec(args)
	if err != nil {
		return impl.Macaroons{}, err
	}
	st, err := mackScheme.PrepareStack(&ms[0], ms[1:])
	if err != nil {
		return impl.Macaroons{}, fmt.Errorf("mack/macaroon.PrepareStack: %w", err)
	}
	return impl.Macaroons{Slice: &st}, nil
}

func (l *Implementation) newMacaroonsFromSpec(args impl.NewMacaroonSpec) ([]macaroon.Macaroon, error) {
	m, err := mackScheme.UnsafeRootMacaroon(args.Location, args.ID, args.RootKey)
	if err != nil {
		return nil, fmt.Errorf("mack/macaroon.UnsafeRootMacaroon: %w", err)
	}
	macaroons := []macaroon.Macaroon{m}
	for _, c := range args.Caveats {
		if c.Location != "" && len(c.Key) != 0 {
			m, err = mackScheme.AddThirdPartyCaveat(&m, c.Key, c.ID, c.Location)
			if err != nil {
				return nil, fmt.Errorf("mack/macaroon.AddThirdPartyCaveat: %w", err)
			}
			macaroons[0] = m
			var discharge []macaroon.Macaroon
			discharge, err = l.newMacaroonsFromSpec(impl.NewMacaroonSpec{
				RootKey:  c.Key,
				ID:       c.ID,
				Location: c.Location,
				Caveats:  c.Caveats,
			})
			if err != nil {
				return nil, err
			}
			macaroons = append(macaroons, discharge...)
			continue
		}
		m, err = mackScheme.AddFirstPartyCaveat(&m, c.ID)
		if err != nil {
			return nil, fmt.Errorf("mack/macaroon.AddFirstPartyCaveat: %w", err)
		}
		macaroons[0] = m
	}
	return macaroons, nil
}

func (l *Implementation) AddFirstPartyCaveat(m impl.Macaroon, cid []byte) (impl.Macaroon, error) {
	in := m.Macaroon.(*macaroon.Macaroon)
	mm, err := mackScheme.AddFirstPartyCaveat(in, cid)
	if err != nil {
		return impl.Macaroon{}, fmt.Errorf("mack/macaroon.AddFirstPartyCaveat: %w", err)
	}
	return impl.Macaroon{Macaroon: &mm}, nil
}

func (l *Implementation) VerifyMacaroon(key []byte, ms impl.Macaroons) (bool, error) {
	st := ms.Slice.(*macaroon.Stack)
	_, err := mackScheme.Verify(context.TODO(), key, *st)
	if err != nil {
		return false, fmt.Errorf("mack/macaroon.Verify: %w", err)
	}
	return true, nil
}

func (l *Implementation) EncodeToV2J(m impl.Macaroon) ([]byte, error) {
	enc := libmacaroon.V2J{}
	m2 := m.Macaroon.(*macaroon.Macaroon)
	bs, err := enc.EncodeMacaroon(m2)
	if err != nil {
		return nil, fmt.Errorf("mack/encoding.V2J.EncodeMacaroon: %w", err)
	}
	return bs, nil
}

func (l *Implementation) EncodeToV2(m impl.Macaroon) ([]byte, error) {
	enc := libmacaroon.V2{}
	m2 := m.Macaroon.(*macaroon.Macaroon)
	bs, err := enc.EncodeMacaroon(m2)
	if err != nil {
		return nil, fmt.Errorf("mack/encoding.V2.EncodeMacaroon: %w", err)
	}
	return bs, nil
}

func (l *Implementation) DecodeFromV2J(bs []byte) (impl.Macaroon, error) {
	enc := libmacaroon.V2J{}
	var m macaroon.Macaroon
	err := enc.DecodeMacaroon(bs, &m)
	if err != nil {
		return impl.Macaroon{}, fmt.Errorf("mack/encoding.V2J.DecodeMacaroon: %w", err)
	}
	return impl.Macaroon{Macaroon: &m}, nil
}

func (l *Implementation) DecodeFromV2(bs []byte) (impl.Macaroon, error) {
	enc := libmacaroon.V2{}
	var m macaroon.Macaroon
	err := enc.DecodeMacaroon(bs, &m)
	if err != nil {
		return impl.Macaroon{}, fmt.Errorf("mack/encoding.V2.DecodeMacaroon: %w", err)
	}
	return impl.Macaroon{Macaroon: &m}, nil
}
