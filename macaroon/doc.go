// Package macaroon contains the core functionality of constructing and verifying a macaroon.
//
// # Create a Scheme
//
// First, you must create a `macaroon.Scheme`. The specification isn't prescriptive about algorithms used,
// so this enables flexibility in the choice of algorithm.
//
// You can create a new scheme with:
//
//	scheme := macaroon.NewScheme(macaroon.SchemeConfig{
//		HMACScheme:           hms,
//		EncryptionScheme:     es,
//		BindForRequestScheme: b4rs,
//	})
//
// Each configuration option is an interface implementing a specific cryptographic algorithm that will be part of the scheme.
//   - [HMACScheme] - Selects Which HMAC algorithm is used to create HMACs.
//   - [EncryptionScheme] - What is used for Encryption/Decryption for Third-Party Caveats (HMAC Size and Encryption Key size must match!)
//   - [BindForRequestScheme] - What function is used to bind a discharge macaroon to an authorization macaroon.
//
// There is a [sensible package] that can be used for creating a [macaroon.Scheme] with sensible defaults:
//
//   - [HMACScheme]: HMAC-SHA256
//   - [EncryptionScheme]: XSalsa20
//   - [BindForRequestScheme] : HMAC-SHA256 where the key is the authorizing macaroon signature
//
// # Create a Macaroon
//
// New Macaroons can be constructed from the Scheme using [Scheme.NewMacaroon].
// You will need a root key, id, location, and initial set of caveats.
//
// Macaroon ID (nonce):
//
//	// A Macaroon ID should be random, and never used again.
//	// UUID might be a good choice.
//	id := make([]byte, 16)
//	random.Read(id)
//
// Macaroon Root Key:
//
//	// You have to be able to associate the Macaroon ID with this key somehow.
//	// maybe store the id/key pair in a DB or construct a key using HMAC with a shared secret and the id?
//	key := make([]byte, scheme.KeySize())
//	random.Read(key)
//
// Macaroon Caveats:
//
//	caveats = [][]byte{
//		[]byte(`org = Organization`),
//		[]byte(`user = User`),
//	}
//	// Creates the initial macaroon.
//	m, _ := scheme.NewMacaroon("https://www.example.com", id, key, caveats...)
//
// You should always have an initial set of at least 1 caveat on a macaroon.
// Without a caveat, a macaroon is permitted to do anything.
// Such a macaroon can be constructed with [Scheme.UnsafeRootMacaroon], but this is not recommended.
//
// # Add First-Party Caveats
//
// First-party caveats are caveats that are cleared by the authorizing service.
//
// The caveat ID is often a predicate that describes a condition that must be true
// if the macaroon is used. How it is parsed, and evaluated is up to the authorizing service,
// so what the bytes represent is opaque to the macaroon scheme.
//
// Additional caveats may be added to a macaroon using `Scheme.AddFirstPartyCaveat`
//
//	m, err = scheme.AddFirstPartyCaveat(&m, []byte(`expires = 2006-01-02`))
//
// # Validating Macaroon Stacks
//
// An Authorizing Macaroon and its associated Discharge Macaroons constitute a [macaroon.Stack].
// Clients construct this stack by receiving discharge macaroons from third-parties for all of their third party caveats
// and presenting both the authorizing macaroon and all discharge macaroons bound to it to the Authorizing Service.
//
// Clients create a stack by using [Scheme.PrepareStack].
//
//	stack, err := scheme.PrepareStack(authorizingMacaroon, dischargeMacaroons)
//
// The stack can then be transmitted with the request. This is also implementation specific, but one way of encoding the stack
// is to use the [encoding/proto] or [encoding/msgpack] packages.
//
// After encoding the stack into bytes, it can be put into a request body or encoded as Base64 and added to an HTTP Authorization
// header. Whatever the service expects.
//
// The Service, after receiving this stack, should decode it (perhaps by using [encoding/proto] or [encoding/msgpack]).
//
// # Verification and Clearing
//
// Validation on the server side happens in two phases: Verifying, and Clearing.
//
// Stack *Verification* only ensures that all the cryptographic signatures match their expected values,
// but it doesn't say anything about the validity of the caveats on the macaroon.
//
//	verifiedStack, err := scheme.Verify(ctx, key, stack)
//
// The key in the above verification function is extracted from the Root Macaroon ID.
// As an example: this could be by looking it up in a database, or by combining the id with some shared secret via an
// HMAC function to generate the key.
//
// After getting a [VerifiedStack] from [Scheme.Verify], the [VerifiedStack] should be cleared before
// allowing the action:
//
//	// checker is a PredicateChecker.
//	// A PredicateChecker interprets a caveat id, and evaluates its result.
//	// It returns true if the predicate is satisfied.
//	if err := verifiedStack.Clear(ctx, checker); err != nil {
//		// stack failed to clear
//		return err
//	}
//	// stack is verified, proceed with action
//
// [sensible package]: github.com/justenwalker/mack/sensible
package macaroon
