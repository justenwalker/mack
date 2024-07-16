# Mack - A Go library for interacting with Macaroons

## What is this library

Currently,

### It is

1. Something I wrote up in my spare time to understand Macaroons a bit better.
2. A learning exercise for me to explore library design ideas.
3. A hobby/side project for me to obsess over.

### It is **not**:

1. A production-grade library you should bet your entire business and information security on.
2. An official or reference implementation of Macaroons. 
3. Able to make Margaritas.

that said, I'd welcome some advice, comments, constructive feedback about the library's, it's ergonomics, or rough edges. I'm also not a cryptographer, so if there are any flaws in the implementation I'd like to know about them.

### Why?

The Macaroons paper specifies how to create this chained list of caveats, but doesn't have any opinions. Unlike JWT, there are really no standards what caveats are, or how to validate them. This library implements the base methods for 
constructing and verifying macaroons and exposes interfaces for implementing the areas of the spec that are open to 
interpretation, while also providing some opinionated implementations of these interfaces to make it actually useful. 

An example implementation of a pair of web-apis using this library can be found in the [example](./example) directory.
This covers constructing a Macaroon, discharging, validating, and clearing caveats in a practical setting.

## What is a Macaroon?

[See Paper](https://www.ndss-symposium.org/wp-content/uploads/2017/09/04_3_1.pdf)

A Macaroon is a security credential that support decentralized delegation.
It is implemented as a chain of signed "Caveats".

Macaroons are similar to [JWT](https://jwt.io) in that they are Bearer tokens
that grant access to a resource. Where they differ is that, JWTs are minted by
a token service with a fixed set of "claims" that cannot be altered; whereas as
Macaroon may derive new sub-macaroons with more limited permissions.

A "caveat" is different from a JWT claim. While
a claim is some (authenticated) information about whom the bearer is, or what the bearer be able to do or what permissions
they have, what groups they are from etc... a caveat only exists as a predicate on what the bearer is allowed to do: it limits
the conditions under which the macaroon is valid.

For example, a list of claims may be:

```json
{
  "user": "foo",
  "groups": ["g1","g2"],
  "permissions": ["p1","p2"]
}
```

Whereas a list of caveats may be
- path /user
- max_size 1MiB


NOTE: some conventional JWT claims can be interpreted as caveats:
- `iss`: token was issuer by this issuer, to limit which token sources should be trusted.
- `aud`: token is valid for a specific audience, preventing token from `service-a` from being used on `service-b`
- `nbf`: token is not valid before this unix timestamp
- `exp`: token is no longer valid after this date

## Packages

- `macaroon` - The main package. These are where all the Macaroon primitive types and operations reside.
- `sensible` - Provides sensible default implementations of cryptographic functions.
- `thirdparty` - Provides a framework for constructing third-party caveats and discharging them.
- `encoding` - Contains a collection of encoding operations on Macaroons for serializing and deserializing them.
- `exchange` - Contains an implementation of the thirdparty macaroons using encrypted caveat ids.

### `macaroon` package

This contains the core functionality of constructing and verifying a macaroon.

#### Create a Macaroon Scheme

First, you must create a `macaroon.Scheme`.

You can create a new scheme with:

```go
scheme := macaroon.NewSceheme(macaroon.SchemeConfig{
	HMACScheme:           hms,
	EncryptionScheme:     es,
	BindForRequestScheme: b4rs,
})
```

Each configuration option is an interface implementing a specific cryptographic algorithm that will be part of the scheme.

- `HMACScheme` - Implements the algorithm is used to generate HMACs.
- `EncryptionScheme` - Implements functions used for Encryption/Decryption for Third-Party Caveats (HMAC Size and Encryption Key size must match!)
- `BindForRequestScheme` - Implements the function used to bind a discharge macaroon to an authorization macaroon.

#### `sensible` package

There is a `sensible` package can be used for creating a `macaroon.Scheme` with sensible defaults:

- `HMACScheme`: HMAC-SHA256
- `EncryptionScheme`: XSalsa20
- `BindForRequestScheme`: discharge.Sig = `HMAC-SHA256(Auth.Sig, Discharge.Sig)`

#### Create a Macaroon

New Macaroons can be constructed from the Scheme using the `NewMacaroon` function:

```go
// start with a scheme, in this case a sensible default
scheme := sensible.Scheme()
	
// Macaroon ID (nonce): Should be random, and never used again.
// - UUID might be a good choice.
id := make([]byte, 16)
random.Read(id)

// Macaroon Root Key:
// You have to be able to associate the Macaroon ID with this key somehow.
// Perhaps you store the id/key pair in a DB or derive a key from a pre-shared password and the macaroon id.
key := make([]byte, scheme.KeySize())
random.Read(key)

// Macaroon Caveats:
// Assemble the list of initial caveats.
// You should always have an initial set of at least 1 caveat on a macaroon.
// Without a caveat, a macaroon is permitted to do anything.
// Such a macaroon can be constructed with Scheme.UnsafeRootMacaroon, but this is not recommended.
caveats = [][]byte{
    []byte(`org = Organization`),
    []byte(`user = User`),
}

// Creates the initial macaroon.
m, err := scheme.NewMacaroon("https://www.example.com", id, key, caveats...)
```
#### Add First-Party Caveats

First-party caveats are caveats that are cleared by the authorizing service.

The caveat ID is typically a predicate that describes a condition that must be true
if the macaroon is used. How it is parsed, and evaluated is up to the authorizing service, 
so what the bytes represent is opaque to the macaroon scheme.

Additional caveats may be added to a macaroon using `Scheme.AddFirstPartyCaveat` 

```go
m, err = scheme.AddFirstPartyCaveat(&m, []byte(`expires = 2006-01-02`))
```


### Validating Macaroon Stacks

An Authorizing Macaroon and its associated Discharge Macaroons constitute a `macaroon.Stack`.
Clients construct this stack by receiving discharge macaroons from third-parties for all of their third party caveats
and presenting both the authorizing macaroon and all discharge macaroons bound to it to the Authorizing Service.

Clients create a stack by using `Scheme.PrepareStack`.

```go
stack, err := scheme.PrepareStack(authorizingMacaroon, dischargeMacaroons)
```

The stack is then transmitted with the request. This is also implementation specific, but one way of encoding the stack
is to use the `encoding/proto` or `encoding/msgpack` packages.

After encoding the stack into bytes, it can be put into a request body or encoded as Base64 and added to an HTTP Authorization
header. Whatever the service expects.

The Service, after receiving this stack, should decode it (perhaps by using `encoding/proto` or `encoding/msgpack`).

Validation on the server side happens in two phases: Verifying, and Clearing.

Stack *Verification* only ensures that all the cryptographic signatures match their expected values,
but it doesn't say anything about the validity of the caveats on the macaroon.

```go
verifiedStack, err := scheme.Verify(ctx, key, stack)
```

The key in the above verification function is extracted from the Root Macaroon ID. As an example: this could be by looking it up
in a database, or by combining the id with some shared secret via an HMAC function to generate the key.

After getting a `VerifiedStack` from the `Scheme.Verify` function, the VerifiedStack should be cleared before
allowing the action:

```go
// checker is a PredicateChecker.
// A PredicateChecker interprets a caveat id, and evaluates its result. 
// It returns true if the predicate is satisfied.
if err := verifiedStack.Clear(ctx, checker); err != nil {
	// stack failed to clear
	return err
}
// stack is verified, proceed with action
// ...
```

### The `macaroon/thirdparty` package

#### (Attenuation) Add Third-Party Caveats to a Macaroon 

While you can construct these values from scratch and use `scheme.AddThirdPartyCaveat`, the `thirdparty` package can help generate 
and apply these values to a Macaroon via a `thirdparty.Attenuator`.

```go
tpa, err := thirdparty.NewAttenuator(thirdparty.AttenuatorConfig{
    Location:     "https://thirparty",
    Scheme:       scheme,
    CaveatIssuer: caveatIssuer,
})
```

- `Location` will be added to each caveat to hint at where to discharge the macaroon later.
- `Scheme` should be provided and match the scheme used to create the macaroon. Mixing schemes will result in undefined behavior.
- `CaveatIDIssuer` issues third party caveat IDs based on a generated caveat key and a predicate evaluated by the third party.

##### Caveat ID Issuer

CaveatIDIssuer is an interface that must be implemented to construct a `thirdparty.Service`. 
It exchanges a `thirdparty.Ticket` which is a pair containing a CaveatKey that is randomly generated for this 
macaroon, and a Predicate to be evaluated by the third party before discharges the caveat; for an opaque caveat ID 
that only the third party can use later to recover the Caveat Key and Predicate.

One way to implement this is by having a third-party implement an API that can take this `thirdparty.Ticket` and 
return a caveat id. This requires the third party to be active in minting a caveat, and typically would create an cId/cK.
Implementing such a protocol is out of scope for this library, but another library implementing
`thirdparty.CaveatIDIssuer` may provide it.

Another way is to use public-key cryptography to encrypt the caveat key/predicate payload. 
This doesn't require a third party to be an active participant in the creation of the caveat.
Instead, the Caveat ID is constructed by encrypting the Caveat Key and Predicate using the third party's public key.

This method implemented in the `exchange` package which can be configured with Encoder/Encryptor implementations for 
which the third party discharge service has a corresponding implementation for Decoder/Decryptor.

```go
issuer := exchange.CaveatIDIssuer{
    Encryptor: encryptor,
    Encoder:   encoder,
}
```
 
Possible implementations for the `Encoder` interface:
- `encoding/proto` - Encodes using [Protobuf v3](https://protobuf.dev/programming-guides/proto3/)
- `encoding/msgpack`- Encodes using [MsgPack](https://msgpack.org/index.html)

A possible implementation for the Encryptor/Decryptor:
- `crypt/agecrypt`: uses [Age](https://age-encryption.org/) to encrypt third-party caveat IDs using Age Recipient.


#### Requesting a Discharge Macaroon from a Third Party

The means by which requests are made to a third party service to create discharge macaroons are not defined
by the spec, but implemented by the end user. This library has some helper to make it easier to discharge
all third-party caveats recursively, returning a collection of discharge macaroons.

As long as the behavior of the discharge request can be described using the ThirdParty interface,
they can be collected into a `thirdparty.Set` which has a Discharge method taking a Macaroon and returning all
the discharge macaroons.

```go
thirdPartySet := thirdparty.Set{authThirdParty}
dischargeMacaroons, err := thirdPartySet.Discharge(ctx, &myMacaroon)
```

#### Discharging Third-Party Caveats

When a Macaroon has third party caveats, they must be discharged to validate the entire macaroon stack.
The bearer of the authorization macaroon should make a request to the third-party service with the caveat ID to discharge
(likely with some authorization for that service too.)

A third-party service may use the `thirdparty.Discharger` to take a CaveatID and return a new Macaroon which will discharge the caveat.

```go
discharger, err := thirdparty.NewDischarger(thirdparty.DischargerConfig{
    Location:         "https://thirparty",
    Scheme:           scheme,
    TicketExtractor:  text,
})
```

- `Location` will be added to the discharge macaroon.
- `Scheme` should be provided and match the scheme used to create the macaroon. Mixing schemes will result in undefined behavior.
- `TicketExtractor` extracts `thirdparty.Ticket` from a caveat ID.

##### TicketExtractor

A TicketExtractor extracts the `thirdparty.Ticket` information from a caveat ID. 
This is the "dual" of the `thirdparty.CaveatIDIssuer`

If the `CaveatIDIssuer` resulted in the third party generating an opaque caveat id and associating it 
to a `thirdparty.Ticket` database, then this would look up that ticket and return it. Implementing such 
a protocol is out of scope for this library, but another library implementing `thirdparty.TicketExtractor` may provide it.

If the CaveatIDIssue instead encrypted the ticket for the third party using its public key, then
this would extract the ticket from the caveat id be decrypting and decoding it.

This method implemented in the `exchange` package which can be configured with Decoder/Decryptor.

```go
text := exchange.TicketExtractor{
    Decryptor: decryptor,
    Decoder:   decoder,
}
```

Possible implementations for the `Decoder` interface:
- `encoding/proto` - Encodes using [Protobuf v3](https://protobuf.dev/programming-guides/proto3/)
- `encoding/msgpack`- Encodes using [MsgPack](https://msgpack.org/index.html)

A possible implementation for the `Decryptor`:
- `crypt/agecrypt`: uses [Age](https://age-encryption.org/) to decrypt caveat IDs using Age Identities.

##### PredicateChecker

A PredicateChecker interprets a caveat id, and evaluates its result. It returns true if the predicate is satisfied.

Once satisfied, the `thirdparty.Discharger` can issue the discharge caveat from the ticket. 
The implementation of the `thirdparty.PredicateChecker` is provided by the user of this library, since
a caveat id is an opaque string of bytes, without any meaning in the macaroon spec.
