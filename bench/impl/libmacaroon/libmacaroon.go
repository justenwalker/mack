package libmacaroon

import (
	"bench/impl"
	"fmt"
	"gopkg.in/macaroon.v2"
)

var _ impl.Interface = (*Implementation)(nil)

type Implementation struct{}

func (l *Implementation) Setup() error {
	return nil
}

func (l *Implementation) NewMacaroon(args impl.NewMacaroonSpec) (impl.Macaroon, error) {
	m, err := macaroon.New(args.RootKey, args.ID, args.Location, macaroon.LatestVersion)
	if err != nil {
		return impl.Macaroon{}, fmt.Errorf("gopkg.in/macaroon.v2.New: %w", err)
	}
	for _, c := range args.Caveats {
		err = m.AddFirstPartyCaveat(c.ID)
		if err != nil {
			return impl.Macaroon{}, fmt.Errorf("gopkg.in/macaroon.v2.AddFirstPartyCaveat: %w", err)
		}
	}
	return impl.Macaroon{Macaroon: m}, nil
}

func (l *Implementation) NewMacaroons(args impl.NewMacaroonSpec) (impl.Macaroons, error) {
	ms, err := l.newMacaroonsFromSpec(args)
	if err != nil {
		return impl.Macaroons{}, fmt.Errorf("gopkg.in/macaroon.v2.newMacaroonsFromSpec: %w", err)
	}
	for _, m := range ms[1:] {
		m.Bind(ms[0].Signature())
	}
	return impl.Macaroons{Slice: ms}, nil
}

func (l *Implementation) newMacaroonsFromSpec(args impl.NewMacaroonSpec) (macaroon.Slice, error) {
	m, err := macaroon.New(args.RootKey, args.ID, args.Location, macaroon.LatestVersion)
	if err != nil {
		return nil, fmt.Errorf("gopkg.in/macaroon.v2.New: %w", err)
	}
	macaroons := macaroon.Slice{m}
	for _, c := range args.Caveats {
		if c.Location != "" && len(c.Key) != 0 {
			err = m.AddThirdPartyCaveat(c.Key, c.ID, c.Location)
			if err != nil {
				return nil, fmt.Errorf("gopkg.in/macaroon.v2.AddThirdPartyCaveat: %w", err)
			}
			var discharge macaroon.Slice
			discharge, err = l.newMacaroonsFromSpec(impl.NewMacaroonSpec{
				RootKey:  c.Key,
				ID:       c.ID,
				Location: c.Location,
				Caveats:  c.Caveats,
			})
			if err != nil {
				return nil, fmt.Errorf("gopkg.in/macaroon.v2.newMacaroonsFromSpec: %w", err)
			}
			macaroons = append(macaroons, discharge...)
			continue
		}
		err = m.AddFirstPartyCaveat(c.ID)
		if err != nil {
			return nil, fmt.Errorf("gopkg.in/macaroon.v2.AddFirstPartyCaveat: %w", err)
		}
	}
	return macaroons, nil
}

func (l *Implementation) AddFirstPartyCaveat(m impl.Macaroon, cid []byte) (impl.Macaroon, error) {
	mine := m.Macaroon.(*macaroon.Macaroon)
	if err := mine.AddFirstPartyCaveat(cid); err != nil {
		return impl.Macaroon{}, fmt.Errorf("gopkg.in/macaroon.v2.AddFirstPartyCaveat: %w", err)
	}
	return impl.Macaroon{Macaroon: mine}, nil
}

func (l *Implementation) VerifyMacaroon(key []byte, ms impl.Macaroons) (bool, error) {
	slice := ms.Slice.(macaroon.Slice)
	err := slice[0].Verify(key, func(caveat string) error {
		return nil
	}, slice[1:])
	if err != nil {
		return false, fmt.Errorf("gopkg.in/macaroon.v2.Verify: %w", err)
	}
	return true, nil
}

func (l *Implementation) EncodeToV2J(m impl.Macaroon) ([]byte, error) {
	mine := m.Macaroon.(*macaroon.Macaroon)
	bs, err := mine.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("gopkg.in/macaroon.v2.Macaroon.MarshalJSON: %w", err)
	}
	return bs, nil
}

func (l *Implementation) EncodeToV2(m impl.Macaroon) ([]byte, error) {
	mine := m.Macaroon.(*macaroon.Macaroon)
	bs, err := mine.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("gopkg.in/macaroon.v2.Macaroon.MarshalJSON: %w", err)
	}
	return bs, nil
}

func (l *Implementation) DecodeFromV2J(bs []byte) (impl.Macaroon, error) {
	var m macaroon.Macaroon
	err := m.UnmarshalJSON(bs)
	if err != nil {
		return impl.Macaroon{}, fmt.Errorf("gopkg.in/macaroon.v2.UnmarshalJSON: %w", err)
	}
	return impl.Macaroon{Macaroon: &m}, nil
}

func (l *Implementation) DecodeFromV2(bs []byte) (impl.Macaroon, error) {
	var m macaroon.Macaroon
	err := m.UnmarshalBinary(bs)
	if err != nil {
		return impl.Macaroon{}, fmt.Errorf("gopkg.in/macaroon.v2.UnmarshalBinary: %w", err)
	}
	return impl.Macaroon{Macaroon: &m}, nil
}
