package macaroon

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/justenwalker/mack/crypt/random"
)

// HMACScheme represents an interface for computing HMAC using a given key and messages.
type HMACScheme interface {
	// KeySize returns the number of bytes required for the HMAC Key.
	// This size should be identical to the Hash used by the HMAC function.
	KeySize() int

	// HMAC computes the HMAC using the given key and messages, writing the output into out.
	// the out byte should have enough capacity receive the full hash output.
	HMAC(key []byte, out []byte, data []byte) error
}

// EncryptionScheme represents an interface for performing encryption and decryption operations.
type EncryptionScheme interface {
	// Overhead returns the additional bytes required for the encrypted payload.
	Overhead() int

	// NonceSize returns the number of bytes required for the Nonce used during encryption.
	NonceSize() int

	// KeySize returns the number of bytes required for the encryption key.
	KeySize() int

	// Encrypt encrypts the plaintext given by `in` and writes the ciphertext to `out`.
	// The in and out []byte buffers should overlap entirely or not at all.
	// The length of the nonce should match the NonceSize() in this scheme.
	// The length of the key should match the KeySize() in this scheme.
	// If `out` is not large enough to contain the ciphertext plus the Overhead, then a new one will be allocated.
	Encrypt(out []byte, in []byte, nonce []byte, key []byte) ([]byte, error)

	// Decrypt decrypts the ciphertext given by `in` and writes the plaintext to `out`.
	// The in and out []byte buffers should overlap entirely or not at all.
	// The length of the key should match the KeySize() in this scheme.
	// If `out` is not large enough to contain the plaintext, then a new one will be allocated.
	Decrypt(out []byte, in []byte, nonce []byte, key []byte) ([]byte, error)
}

// BindForRequestScheme represents an interface for binding a third-party discharge Macaroon for a request.
type BindForRequestScheme interface {
	// BindForRequest takes a Macaroon for the target service and the signature sig of the discharge Macaroon.
	// It overwrites the bytes in sig with a new signature, binding it to the target service Macaroon provided.
	BindForRequest(ts *Macaroon, sig []byte) error
}

// SchemeConfig configures a [NewScheme].
// It contains the set of algorithms used in constructing a [Macaroon].
type SchemeConfig struct {
	// HMACScheme is the implementation of the HMAC algorithm is used to create HMACs.
	HMACScheme HMACScheme
	// EncryptionScheme is the implementation of Encryption/Decryption for Third-Party Caveats
	// The HMAC Key Size and Encryption Key size must match.
	EncryptionScheme EncryptionScheme
	// BindForRequestScheme is the implementation for binding a discharge macaroon to an authorization macaroon.
	BindForRequestScheme BindForRequestScheme
}

// NewScheme creates a new macaroon scheme from the given [SchemeConfig].
// A [Scheme] is the primary way that a [Macaroon] is constructed, modified, and verified
// since a macaroon itself contains no data about what algorithms were used in its construction.
func NewScheme(cfg SchemeConfig) (*Scheme, error) {
	if cfg.HMACScheme == nil {
		return nil, errors.New("NewScheme: HMACScheme must be provided")
	}
	if cfg.EncryptionScheme == nil {
		return nil, errors.New("NewScheme: EncryptionScheme must be provided")
	}
	if cfg.BindForRequestScheme == nil {
		return nil, errors.New("NewScheme: BindForRequestScheme must be provided")
	}
	if cfg.HMACScheme.KeySize() != cfg.EncryptionScheme.KeySize() {
		return nil, fmt.Errorf("NewScheme: KeySize : HMACScheme.KeySize=%d, EncryptionScheme.KeySize=%d", cfg.HMACScheme.KeySize(), cfg.EncryptionScheme.KeySize())
	}
	keySize := cfg.HMACScheme.KeySize()
	return &Scheme{
		hmac:      cfg.HMACScheme,
		enc:       cfg.EncryptionScheme,
		bfr:       cfg.BindForRequestScheme,
		keySize:   cfg.EncryptionScheme.KeySize(),
		nonceSize: cfg.EncryptionScheme.NonceSize(),
		overhead:  cfg.EncryptionScheme.Overhead(),

		keyPool: &sync.Pool{
			New: func() interface{} {
				bs := make([]byte, keySize)
				return &bs
			},
		},
	}, nil
}

// Scheme implements the common cryptographic routines necessary to create, modify, and verify Macaroons.
type Scheme struct {
	hmac      HMACScheme
	enc       EncryptionScheme
	bfr       BindForRequestScheme
	keySize   int
	nonceSize int
	overhead  int

	// keyPool helps optimize the third-party caveat verification process by eliminating allocations
	keyPool *sync.Pool
}

// UnsafeRootMacaroon creates a new root-level Macaroon with the given id and key.
// This operation is typically unsafe since a Root macaroon has no caveats, and this can authorize any action.
// Use [Scheme.NewMacaroon] instead which takes a mandatory caveat.
func (s *Scheme) UnsafeRootMacaroon(loc string, id []byte, key []byte) (Macaroon, error) {
	return s.newMacaroon(loc, id, key)
}

// NewMacaroon creates a new Macaroon using the given key and ID.
// This operation requires at least one caveat to be provided, and will fail is it is empty.
// Without a caveat, a macaroon is permitted to do anything.
// Such a macaroon can be constructed with [Scheme.UnsafeRootMacaroon], but this is not recommended.
func (s *Scheme) NewMacaroon(loc string, id []byte, key []byte, caveats ...[]byte) (Macaroon, error) {
	if len(caveats) == 0 {
		return Macaroon{}, errors.New("at least one caveat must be provided")
	}
	for i := range caveats {
		if len(caveats[i]) == 0 {
			return Macaroon{}, errors.New("empty caveats are invalid")
		}
	}
	m, err := s.newMacaroon(loc, id, key)
	if err != nil {
		return m, err
	}
	rcs := make([]RawCaveat, len(caveats))
	for i, c := range caveats {
		rcs[i] = RawCaveat{
			CID: c,
		}
	}
	m = m.addCaveats(s, rcs...)
	return m, nil
}

func (s *Scheme) newMacaroon(loc string, id []byte, key []byte) (Macaroon, error) {
	if len(key) != s.keySize {
		return Macaroon{}, fmt.Errorf("%w: invalid key size. need=%d, got=%d", ErrInvalidArgument, s.keySize, len(key))
	}
	if len(id) == 0 {
		return Macaroon{}, fmt.Errorf("%w: macaroon id cannot be empty", ErrInvalidArgument)
	}
	keyBuf := s.getKeyBuffer()
	copy(*keyBuf, key)
	defer s.releaseKeyBuffer(keyBuf)
	return newMacaroon(s, *keyBuf, id, loc)
}

// KeySize returns the length of the macaroon HMAC and Encryption keys in bytes.
func (s *Scheme) KeySize() int {
	return s.keySize
}

// Verify the cryptographic signatures of the entire macaroon stack using the root key provided.
func (s *Scheme) Verify(ctx context.Context, key []byte, stack Stack) (VerifiedStack, error) {
	v := getVerifyContext(ctx)
	v.init(stack)
	if len(key) != s.keySize {
		return VerifiedStack{}, fmt.Errorf("%w: invalid key size. need=%d, got=%d", ErrInvalidArgument, s.keySize, len(key))
	}
	target := &stack[0]
	discharge := stack[1:]
	var discharged []byte
	if len(discharge) > 32 {
		discharged = make([]byte, len(discharge))
	} else {
		var ds [32]byte
		discharged = ds[:len(discharge)]
	}
	keyBuf := s.getKeyBuffer()
	copy(*keyBuf, key)
	defer s.releaseKeyBuffer(keyBuf)
	if err := target.verify(s, stack, *keyBuf, *keyBuf, v, 0, discharged); err != nil {
		return VerifiedStack{}, err
	}
	for i := range discharged {
		if discharged[i] == 0 {
			err := validationError(target, fmt.Errorf("discharge macaroon %d was unused", i))
			v.fail(0, err)
			return VerifiedStack{}, err
		}
		if discharged[i] > 1 {
			err := validationError(target, fmt.Errorf("discharge macaroon %d was used more than once", i))
			v.fail(0, err)
			return VerifiedStack{}, err
		}
	}
	return VerifiedStack{
		stack: stack,
	}, nil
}

// PrepareStack prepares the set of discharge macaroons for a request, assembling them with the target Macaroon into a [Stack].
// This stack can be presented to the authorizing service for verification.
func (s *Scheme) PrepareStack(m *Macaroon, discharge []Macaroon) (Stack, error) {
	stack := make([]Macaroon, 0, 1+len(discharge))
	stack = append(stack, *m)
	for i := range discharge {
		mu, err := s.BindForRequest(m, &discharge[i])
		if err != nil {
			return Stack{}, err
		}
		stack = append(stack, mu)
	}
	return stack, nil
}

// AddFirstPartyCaveat appends a first party caveat, returning a new Macaroon with the caveat appended.
func (s *Scheme) AddFirstPartyCaveat(m *Macaroon, cID []byte) (Macaroon, error) {
	if len(cID) == 0 {
		return Macaroon{}, errors.New("AddFirstPartyCaveat: empty predicate")
	}
	return m.addFirstPartyCaveat(s, cID), nil
}

// AddThirdPartyCaveat appends a third party caveat, returning a new Macaroon with the caveat appended.
// Coordinating the link between the cKey and cID is out of scope for this function.
func (s *Scheme) AddThirdPartyCaveat(m *Macaroon, cKey []byte, cID []byte, location string) (Macaroon, error) {
	if len(cKey) == 0 {
		return Macaroon{}, errors.New("AddThirdPartyCaveat: empty key (cK)")
	}
	if len(cID) == 0 {
		return Macaroon{}, errors.New("AddThirdPartyCaveat: empty caveat id (cID)")
	}
	return m.addThirdPartyCaveat(s, cKey, cID, location)
}

// BindForRequest creates a new Discharge Macaroon that uses.
func (s *Scheme) BindForRequest(target *Macaroon, discharge *Macaroon) (Macaroon, error) {
	c := Clone(discharge)
	if err := c.bindForRequest(s, target); err != nil {
		return Macaroon{}, err
	}
	return c, nil
}

func (s *Scheme) encrypt(out []byte, in []byte, key []byte) ([]byte, error) {
	if len(key) != s.keySize {
		return nil, fmt.Errorf("%w: invalid key size. need=%d, got=%d", ErrInvalidArgument, s.keySize, len(key))
	}
	out = s.growBuffer(out, len(in)+s.overhead+s.nonceSize)
	random.Read(out[cap(out)-s.nonceSize:])

	_, err := s.enc.Encrypt(out[:len(in)+s.overhead], in, out[cap(out)-s.nonceSize:], key)
	if err != nil {
		return nil, fmt.Errorf("macaroon: error encrypting: %w", err)
	}
	return out[:len(in)+s.overhead+s.nonceSize], nil
}

func (s *Scheme) decrypt(out []byte, in []byte, key []byte) ([]byte, error) {
	if len(in) < s.nonceSize+s.overhead {
		return nil, fmt.Errorf("%w: input buffer too small. need=%d, got=%d", ErrInvalidArgument, s.nonceSize+s.overhead, len(in))
	}
	if len(key) != s.keySize {
		return nil, fmt.Errorf("%w: invalid key size. need=%d, got=%d", ErrInvalidArgument, s.keySize, len(key))
	}
	out = s.growBuffer(out, len(in)-s.overhead-s.nonceSize)
	nonce := in[len(in)-s.nonceSize:]
	ciphertext := in[:len(in)-s.nonceSize]
	return s.enc.Decrypt(out[:len(in)-s.overhead-s.nonceSize], ciphertext, nonce, key)
}

func (s *Scheme) growBuffer(buf []byte, sz int) []byte {
	if cap(buf) < sz {
		n := make([]byte, len(buf), sz)
		copy(n, buf)
		buf = n
	}
	return buf[:sz:sz]
}

func (s *Scheme) getKeyBuffer() *[]byte {
	v := s.keyPool.Get()
	return v.(*[]byte) //nolint:forcetypeassert
}

func (s *Scheme) releaseKeyBuffer(p *[]byte) {
	// zero memory before putting the buffer back in the pool, so we don't leak key material between operations.
	// this should optimize to memclr; see: https://github.com/golang/go/issues/5373
	for i := range *p {
		(*p)[i] = 0
	}
	s.keyPool.Put(p)
}
