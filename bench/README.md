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

| Hardware    | ID             | Model #   | CPU          | Cores      | Memory |
|-------------|----------------|-----------|--------------|------------|--------|
| MacBook Pro | MacBookPro18,3 | MKGP3LL/A | Apple M1 Pro | proc 8:6:2 | 16 GB  |

## Results

```
goos: darwin
goarch: arm64
pkg: bench
                      │ libmacaroon  │                mack                 │
                      │    sec/op    │   sec/op     vs base                │
NewMacaroon-8            898.0n ± 1%   405.8n ± 0%  -54.81% (p=0.000 n=10)
AddFirstPartyCaveat-8    473.5n ± 1%   351.6n ± 0%  -25.76% (p=0.000 n=10)
Verify_small-8          1055.0n ± 1%   431.1n ± 0%  -59.14% (p=0.000 n=10)
Verify_large-8          20.016µ ± 0%   6.765µ ± 0%  -66.20% (p=0.000 n=10)
EncodeToV2J-8            753.1n ± 1%   631.4n ± 1%  -16.17% (p=0.000 n=10)
EncodeToV2-8             195.5n ± 0%   172.8n ± 1%  -11.59% (p=0.000 n=10)
DecodeFromV2J-8          1.748µ ± 0%   2.277µ ± 1%  +30.23% (p=0.000 n=10)
DecodeFromV2-8           238.3n ± 1%   399.2n ± 1%  +67.51% (p=0.000 n=10)
geomean                  928.2n        671.4n       -27.67%

                      │ libmacaroon  │                   mack                    │
                      │     B/op     │     B/op      vs base                     │
NewMacaroon-8            1472.0 ± 0%     344.0 ± 0%   -76.63% (p=0.000 n=10)
AddFirstPartyCaveat-8     866.0 ± 1%     432.0 ± 0%   -50.12% (p=0.000 n=10)
Verify_small-8          1.508Ki ± 0%   0.000Ki ± 0%  -100.00% (p=0.000 n=10)
Verify_large-8          26.20Ki ± 0%    0.00Ki ± 0%  -100.00% (p=0.000 n=10)
EncodeToV2J-8             824.0 ± 0%     712.0 ± 0%   -13.59% (p=0.000 n=10)
EncodeToV2-8              720.0 ± 0%     336.0 ± 0%   -53.33% (p=0.000 n=10)
DecodeFromV2J-8         1.023Ki ± 0%   1.953Ki ± 0%   +90.84% (p=0.000 n=10)
DecodeFromV2-8            672.0 ± 0%     674.0 ± 0%    +0.30% (p=0.000 n=10)
geomean                 1.438Ki                      ?                       ¹ ²
¹ summaries must be >0 to compute geomean
² ratios must be >0 to compute geomean

                      │ libmacaroon │                   mack                   │
                      │  allocs/op  │  allocs/op   vs base                     │
NewMacaroon-8           18.000 ± 0%    5.000 ± 0%   -72.22% (p=0.000 n=10)
AddFirstPartyCaveat-8    7.000 ± 0%    3.000 ± 0%   -57.14% (p=0.000 n=10)
Verify_small-8           19.00 ± 0%     0.00 ± 0%  -100.00% (p=0.000 n=10)
Verify_large-8           331.0 ± 0%      0.0 ± 0%  -100.00% (p=0.000 n=10)
EncodeToV2J-8            8.000 ± 0%    7.000 ± 0%   -12.50% (p=0.000 n=10)
EncodeToV2-8             5.000 ± 0%    4.000 ± 0%   -20.00% (p=0.000 n=10)
DecodeFromV2J-8          14.00 ± 0%    21.00 ± 0%   +50.00% (p=0.000 n=10)
DecodeFromV2-8           7.000 ± 0%   19.000 ± 0%  +171.43% (p=0.000 n=10)
geomean                  15.36                     ?                       ¹ ²
¹ summaries must be >0 to compute geomean
² ratios must be >0 to compute geomean
```
