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

## Hardware

| Hardware     | ID             | Model #    | CPU           | Cores       | Memory |
|--------------|----------------|------------|---------------|-------------|--------|
| MacBook Pro  | MacBookPro18,3 | MKGP3LL/A  | Apple M1 Pro  | proc 8:6:2  | 16 GB  |

## Results

```
goos: darwin
goarch: arm64
pkg: bench
cpu: Apple M1 Pro
                      │ libmacaroon  │                mack                 │
                      │    sec/op    │   sec/op     vs base                │
NewMacaroon-8            816.6n ± 1%   354.6n ± 0%  -56.58% (p=0.000 n=10)
AddFirstPartyCaveat-8    466.6n ± 2%   308.9n ± 0%  -33.80% (p=0.000 n=10)
Verify_small-8          1016.0n ± 1%   416.7n ± 0%  -58.99% (p=0.000 n=10)
Verify_large-8          19.151µ ± 1%   7.356µ ± 0%  -61.59% (p=0.000 n=10)
EncodeToV2J-8            683.9n ± 1%   575.9n ± 1%  -15.78% (p=0.000 n=10)
EncodeToV2-8             157.1n ± 1%   137.2n ± 0%  -12.69% (p=0.000 n=10)
DecodeFromV2J-8          1.754µ ± 1%   1.961µ ± 0%  +11.83% (p=0.000 n=10)
DecodeFromV2-8           188.2n ± 1%   133.3n ± 1%  -29.16% (p=0.000 n=10)
geomean                  846.3n        537.2n       -36.52%

                      │  libmacaroon  │                   mack                    │
                      │     B/op      │     B/op      vs base                     │
NewMacaroon-8             1456.0 ± 0%     328.0 ± 0%   -77.47% (p=0.000 n=10)
AddFirstPartyCaveat-8      896.0 ± 7%     416.0 ± 0%   -53.57% (p=0.000 n=10)
Verify_small-8           1.508Ki ± 0%   0.000Ki ± 0%  -100.00% (p=0.000 n=10)
Verify_large-8          26.203Ki ± 0%   7.503Ki ± 0%   -71.37% (p=0.000 n=10)
EncodeToV2J-8              800.0 ± 0%     688.0 ± 0%   -14.00% (p=0.000 n=10)
EncodeToV2-8               696.0 ± 0%     312.0 ± 0%   -55.17% (p=0.000 n=10)
DecodeFromV2J-8          1.039Ki ± 0%   1.195Ki ± 0%   +15.04% (p=0.000 n=10)
DecodeFromV2-8             656.0 ± 0%     368.0 ± 0%   -43.90% (p=0.000 n=10)
geomean                  1.429Ki                      ?                       ¹ ²
¹ summaries must be >0 to compute geomean
² ratios must be >0 to compute geomean

                      │ libmacaroon │                  mack                   │
                      │  allocs/op  │ allocs/op   vs base                     │
NewMacaroon-8           17.000 ± 0%   4.000 ± 0%   -76.47% (p=0.000 n=10)
AddFirstPartyCaveat-8    6.000 ± 0%   2.000 ± 0%   -66.67% (p=0.000 n=10)
Verify_small-8           19.00 ± 0%    0.00 ± 0%  -100.00% (p=0.000 n=10)
Verify_large-8          331.00 ± 0%   12.00 ± 0%   -96.37% (p=0.000 n=10)
EncodeToV2J-8            7.000 ± 0%   6.000 ± 0%   -14.29% (p=0.000 n=10)
EncodeToV2-8             4.000 ± 0%   3.000 ± 0%   -25.00% (p=0.000 n=10)
DecodeFromV2J-8          14.00 ± 0%   16.00 ± 0%   +14.29% (p=0.000 n=10)
DecodeFromV2-8           6.000 ± 0%   4.000 ± 0%   -33.33% (p=0.000 n=10)
geomean                  14.04                    ?                       ¹ ²
¹ summaries must be >0 to compute geomean
² ratios must be >0 to compute geomean
```
