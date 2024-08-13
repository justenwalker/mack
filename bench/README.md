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

| Hardware     | ID             | Model #   | CPU          | Cores      | Memory |
|--------------|----------------|-----------|--------------|------------|--------|
| MacBook Pro  | MacBookPro18,3 | MKGP3LL/A | Apple M1 Pro | proc 8:6:2 | 16 GB  |

## Results

```
goos: darwin
goarch: arm64
pkg: bench
cpu: Apple M1 Pro
                      │ libmacaroon  │                mack                 │
                      │    sec/op    │   sec/op     vs base                │
NewMacaroon-8            817.7n ± 6%   357.8n ± 3%  -56.24% (p=0.000 n=10)
AddFirstPartyCaveat-8    467.9n ± 4%   308.4n ± 0%  -34.11% (p=0.000 n=10)
Verify_small-8          1015.5n ± 1%   430.5n ± 1%  -57.61% (p=0.000 n=10)
Verify_large-8          19.224µ ± 1%   7.402µ ± 0%  -61.50% (p=0.000 n=10)
EncodeToV2J-8            687.8n ± 0%   576.9n ± 0%  -16.13% (p=0.000 n=10)
EncodeToV2-8             156.7n ± 1%   137.1n ± 0%  -12.45% (p=0.000 n=10)
DecodeFromV2J-8          1.754µ ± 0%   2.159µ ± 0%  +23.06% (p=0.000 n=10)
DecodeFromV2-8           189.9n ± 1%   269.1n ± 1%  +41.74% (p=0.000 n=10)
geomean                  848.2n        597.1n       -29.61%

                      │  libmacaroon  │                   mack                    │
                      │     B/op      │     B/op      vs base                     │
NewMacaroon-8             1456.0 ± 0%     328.0 ± 0%   -77.47% (p=0.000 n=10)
AddFirstPartyCaveat-8      900.0 ± 3%     416.0 ± 0%   -53.78% (p=0.000 n=10)
Verify_small-8           1.508Ki ± 0%   0.000Ki ± 0%  -100.00% (p=0.000 n=10)
Verify_large-8          26.203Ki ± 0%   7.503Ki ± 0%   -71.37% (p=0.000 n=10)
EncodeToV2J-8              800.0 ± 0%     688.0 ± 0%   -14.00% (p=0.000 n=10)
EncodeToV2-8               696.0 ± 0%     312.0 ± 0%   -55.17% (p=0.000 n=10)
DecodeFromV2J-8          1.039Ki ± 0%   1.938Ki ± 0%   +86.47% (p=0.000 n=10)
DecodeFromV2-8             656.0 ± 0%     656.0 ± 0%         ~ (p=1.000 n=10) ¹
geomean                  1.430Ki                      ?                       ² ³
¹ all samples are equal
² summaries must be >0 to compute geomean
³ ratios must be >0 to compute geomean

                      │ libmacaroon │                  mack                   │
                      │  allocs/op  │ allocs/op   vs base                     │
NewMacaroon-8           17.000 ± 0%   4.000 ± 0%   -76.47% (p=0.000 n=10)
AddFirstPartyCaveat-8    6.000 ± 0%   2.000 ± 0%   -66.67% (p=0.000 n=10)
Verify_small-8           19.00 ± 0%    0.00 ± 0%  -100.00% (p=0.000 n=10)
Verify_large-8          331.00 ± 0%   12.00 ± 0%   -96.37% (p=0.000 n=10)
EncodeToV2J-8            7.000 ± 0%   6.000 ± 0%   -14.29% (p=0.000 n=10)
EncodeToV2-8             4.000 ± 0%   3.000 ± 0%   -25.00% (p=0.000 n=10)
DecodeFromV2J-8          14.00 ± 0%   20.00 ± 0%   +42.86% (p=0.000 n=10)
DecodeFromV2-8           6.000 ± 0%   8.000 ± 0%   +33.33% (p=0.000 n=10)
geomean                  14.04                    ?                       ¹ ²
¹ summaries must be >0 to compute geomean
² ratios must be >0 to compute geomean
```
