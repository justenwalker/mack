# Benchmarks

## Hardware

| Hardware    | ID             | Model #    | CPU          | Cores     | Memory |
|-------------|----------------|------------|--------------|-----------|--------|
| MacBook Pro | MacBookPro18,3 | MKGP3LL/A  | Apple M1 Pro | 8 (6p/2e) | 16 GB  |

## Tests

| Test                     | Description                                                                      |
|--------------------------|----------------------------------------------------------------------------------|
| **NewMacaroon**          | Create a small macaroon with just the root key and id                            |
| **AddFirstPartyCaveat**  | Add a 100 random byte caveat to an existing macaroon                             |
| **Verify_small**         | Verify a small macaroon with a single caveat                                     |
| **Verify_large**         | Verify a large, multi-level macaroon with several nested third-party discharges  |
| **EncodeToV2J**          | Encode a small macaroon into JSON using libmacaroon/v2j format                   |
| **EncodeToV2**           | Encode a small macaroon into binary using libmacaroon/v2 format                  |

## Results

```
goos: darwin
goarch: arm64
pkg: bench
                      │ libmacaroon  │                mack                 │
                      │    sec/op    │   sec/op     vs base                │
NewMacaroon-8            902.2n ± 0%   404.5n ± 0%  -55.16% (p=0.000 n=10)
AddFirstPartyCaveat-8    487.4n ± 1%   350.7n ± 0%  -28.06% (p=0.000 n=10)
Verify_small-8          1066.0n ± 0%   432.1n ± 0%  -59.47% (p=0.000 n=10)
Verify_large-8          20.176µ ± 0%   6.769µ ± 0%  -66.45% (p=0.000 n=10)
EncodeToV2J-8            750.1n ± 1%   689.5n ± 0%   -8.07% (p=0.000 n=10)
EncodeToV2-8             195.5n ± 1%   303.0n ± 0%  +55.03% (p=0.000 n=10)
geomean                  1.056µ        665.2n       -37.00%

                      │ libmacaroon  │                   mack                    │
                      │     B/op     │     B/op      vs base                     │
NewMacaroon-8            1472.0 ± 0%     344.0 ± 0%   -76.63% (p=0.000 n=10)
AddFirstPartyCaveat-8     875.5 ± 1%     432.0 ± 0%   -50.66% (p=0.000 n=10)
Verify_small-8          1.508Ki ± 0%   0.000Ki ± 0%  -100.00% (p=0.000 n=10)
Verify_large-8          26.20Ki ± 0%    0.00Ki ± 0%  -100.00% (p=0.000 n=10)
EncodeToV2J-8             824.0 ± 0%     776.0 ± 0%    -5.83% (p=0.000 n=10)
EncodeToV2-8              720.0 ± 0%     616.0 ± 0%   -14.44% (p=0.000 n=10)
geomean                 1.737Ki                      ?                       ¹ ²
¹ summaries must be >0 to compute geomean
² ratios must be >0 to compute geomean

                      │ libmacaroon │                   mack                   │
                      │  allocs/op  │  allocs/op   vs base                     │
NewMacaroon-8           18.000 ± 0%    5.000 ± 0%   -72.22% (p=0.000 n=10)
AddFirstPartyCaveat-8    7.000 ± 0%    3.000 ± 0%   -57.14% (p=0.000 n=10)
Verify_small-8           19.00 ± 0%     0.00 ± 0%  -100.00% (p=0.000 n=10)
Verify_large-8           331.0 ± 0%      0.0 ± 0%  -100.00% (p=0.000 n=10)
EncodeToV2J-8            8.000 ± 0%   10.000 ± 0%   +25.00% (p=0.000 n=10)
EncodeToV2-8             5.000 ± 0%   16.000 ± 0%  +220.00% (p=0.000 n=10)
geomean                  17.79                     ?                       ¹ ²
¹ summaries must be >0 to compute geomean
² ratios must be >0 to compute geomean
```
