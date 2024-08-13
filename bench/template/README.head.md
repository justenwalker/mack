# Benchmarks

## Tests

| Test                     | Description                                                                     |
|--------------------------|---------------------------------------------------------------------------------|
| **NewMacaroon**          | Create a small macaroon with just the root key and id                           |
| **AddFirstPartyCaveat**  | Add a 100 random byte caveat to an existing macaroon                            |
| **Verify_small**         | Verify a small macaroon with a single caveat                                    |
| **Verify_large**         | Verify a large, multi-level macaroon with several nested third-party discharges |
| **EncodeToV2J**          | Encode a small macaroon into JSON using libmacaroon/v2j format                  |
| **EncodeToV2**           | Encode a small macaroon into binary using libmacaroon/v2 format                 |
| **DecodeFromV2J**        | Decode a small macaroon from JSON using libmacaroon/v2j format                  |
| **DecodeFromV2**         | Decode a small macaroon from binary using libmacaroon/v2 format                 |

