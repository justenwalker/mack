package impl

type Interface interface {
	Setup() error
	NewMacaroon(args NewMacaroonSpec) (Macaroon, error)
	NewMacaroons(args NewMacaroonSpec) (Macaroons, error)
	AddFirstPartyCaveat(m Macaroon, cid []byte) (Macaroon, error)
	VerifyMacaroon(key []byte, ms Macaroons) (bool, error)
	EncodeToV2J(m Macaroon) ([]byte, error)
	EncodeToV2(m Macaroon) ([]byte, error)
	DecodeFromV2J(bs []byte) (Macaroon, error)
	DecodeFromV2(bs []byte) (Macaroon, error)
}

type NewMacaroonSpec struct {
	// RootKey is the root key of the macaroon.
	RootKey []byte
	// ID is the macaroon ID or nonce that will be signed by the root key.
	ID []byte
	// Location is the macaroon location string.
	Location string
	// Caveats contain a collection of zero or more caveats to apply to the macaroon.
	Caveats []NewCaveatSpec
}

type (
	// Macaroons encapsulates one or more macaroon objects from the target library.
	Macaroons struct {
		Slice interface{}
	}
	// Macaroon represents a single macaroon object from the target library implementation.
	Macaroon struct {
		Macaroon interface{}
	}
)

type NewCaveatSpec struct {
	// ID is a mandatory caveat id.
	ID []byte

	// Key is the caveat key for a third party caveat. If set, this indicates it is a third-party caveat.
	Key []byte
	// Location is the location of the caveat. this must be set if a caveat key is given.
	Location string

	// Caveats are optional additional caveats to apply to a discharge macaroon for this third party caveat.
	Caveats []NewCaveatSpec
}

type Implementation struct {
	Name string
	Interface
}
