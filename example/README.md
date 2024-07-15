# Example

This module is an example of a pair of services implementing Macaroons for authorization for API operations.

- `auth` - contains the authentication service, which issues discharge macaroons authorizing requests
- `target` - contains the target service, which authorizes operations using macaroons, and creates new macaroons
  from auth tokens obtained from the auth service.

The main packager starts these services, and creates a macaroon with a third-party caveat which is discharged by the auth service.

1. It tries a request for which it has appropriate access, and succeed.
2. It tries a request for which is unauthorized, and it fails.
3. It sends a macaroon that has a bad signature, and it fails.

## Output

```
2024/07/11 00:00:00 Starting API Servers
2024/07/11 00:00:00 - Acquire Macaroon
2024/07/11 00:00:00  => Macaroon: {
  "location": "http://127.0.0.1:8081",
  "id": [
    "~hex",
    "cccdfc9d3c21489abf5dca73d56c88c5"
  ],
  "caveats": [
    {
      "cid": "{\"op\":\"org\",\"args\":[\"myorg\"]}"
    },
    {
      "cid": "{\"op\":\"app\",\"args\":[\"myapp\"]}"
    },
    {
      "cid": "{\"op\":\"expires\",\"args\":[\"2024-07-16T07:56:58-04:00\"]}"
    },
    {
      "location": "http://127.0.0.1:8081",
      "vid": [
        "~hex",
        "3c60eba1f59af27b29355651997e79504c9f20f9fc0751d830ee7c27652cdc1f10b1d75f4fa15d508736b822f0516f3303fdaf7020850ba9"
      ],
      "cid": [
        "~hex",
        "83a474797065a66167653a7631a36b6964a46b696431a464617461c5010f6167652d656e6372797074696f6e2e6f72672f76310a2d3e205832353531392034374e6c3754476c4a61455357486374524f456d43397351374e474d2b757656744d626d5859693053786b0a78445a573749645a74466d7a632f41562b6a41593253654265306249432b7052496e6c6b597a65444750630a2d2d2d2045324f6e394a5878472f387266466a4c6956754a50523142563956672b50467976617331654467483245490a77c3f917fdb820cd047682a1398aa14215a1da0697feeec07d06fd29664e3c523f8a1b0aad1de7e4748baef914d19dfcdf3cc22d68e90239b105abd46eaf809450002c75f198ad0b3103e6c78b0daad495fb2f4817a04a7e937d73c16f15ed6d57a2b9d9d842d3"
      ]
    }
  ],
  "sig": [
    "~hex",
    "e5cc362e341d84bd24840e90e75ad7c56a409c9afb238e7fd08b1877602d0f20"
  ]
}

2024/07/11 00:00:00 - Discharge Macaroon Caveats
2024/07/11 00:00:00  => Discharge[0]: {
  "location": "http://127.0.0.1:8080",
  "id": [
    "~hex",
    "83a474797065a66167653a7631a36b6964a46b696431a464617461c5010f6167652d656e6372797074696f6e2e6f72672f76310a2d3e205832353531392034374e6c3754476c4a61455357486374524f456d43397351374e474d2b757656744d626d5859693053786b0a78445a573749645a74466d7a632f41562b6a41593253654265306249432b7052496e6c6b597a65444750630a2d2d2d2045324f6e394a5878472f387266466a4c6956754a50523142563956672b50467976617331654467483245490a77c3f917fdb820cd047682a1398aa14215a1da0697feeec07d06fd29664e3c523f8a1b0aad1de7e4748baef914d19dfcdf3cc22d68e90239b105abd46eaf809450002c75f198ad0b3103e6c78b0daad495fb2f4817a04a7e937d73c16f15ed6d57a2b9d9d842d3"
  ],
  "caveats": [
    {
      "cid": "{\"op\":\"expires\",\"args\":[\"2024-07-16T04:01:58Z\"]}"
    }
  ],
  "sig": [
    "~hex",
    "8b3759d3c9c2d046d5b52a5e4a9e39b01944ecc6ea44d3752e5a1093a4bdddcf"
  ]
}
2024/07/11 00:00:00 - Preparing Macaroon Stack
2024/07/11 00:00:00  => Stack[0]: {
  "location": "http://127.0.0.1:8081",
  "id": [
    "~hex",
    "cccdfc9d3c21489abf5dca73d56c88c5"
  ],
  "caveats": [
    {
      "cid": "{\"op\":\"org\",\"args\":[\"myorg\"]}"
    },
    {
      "cid": "{\"op\":\"app\",\"args\":[\"myapp\"]}"
    },
    {
      "cid": "{\"op\":\"expires\",\"args\":[\"2024-07-16T07:56:58-04:00\"]}"
    },
    {
      "location": "http://127.0.0.1:8081",
      "vid": [
        "~hex",
        "3c60eba1f59af27b29355651997e79504c9f20f9fc0751d830ee7c27652cdc1f10b1d75f4fa15d508736b822f0516f3303fdaf7020850ba9"
      ],
      "cid": [
        "~hex",
        "83a474797065a66167653a7631a36b6964a46b696431a464617461c5010f6167652d656e6372797074696f6e2e6f72672f76310a2d3e205832353531392034374e6c3754476c4a61455357486374524f456d43397351374e474d2b757656744d626d5859693053786b0a78445a573749645a74466d7a632f41562b6a41593253654265306249432b7052496e6c6b597a65444750630a2d2d2d2045324f6e394a5878472f387266466a4c6956754a50523142563956672b50467976617331654467483245490a77c3f917fdb820cd047682a1398aa14215a1da0697feeec07d06fd29664e3c523f8a1b0aad1de7e4748baef914d19dfcdf3cc22d68e90239b105abd46eaf809450002c75f198ad0b3103e6c78b0daad495fb2f4817a04a7e937d73c16f15ed6d57a2b9d9d842d3"
      ]
    }
  ],
  "sig": [
    "~hex",
    "e5cc362e341d84bd24840e90e75ad7c56a409c9afb238e7fd08b1877602d0f20"
  ]
}
2024/07/11 00:00:00  => Stack[1]: {
  "location": "http://127.0.0.1:8080",
  "id": [
    "~hex",
    "83a474797065a66167653a7631a36b6964a46b696431a464617461c5010f6167652d656e6372797074696f6e2e6f72672f76310a2d3e205832353531392034374e6c3754476c4a61455357486374524f456d43397351374e474d2b757656744d626d5859693053786b0a78445a573749645a74466d7a632f41562b6a41593253654265306249432b7052496e6c6b597a65444750630a2d2d2d2045324f6e394a5878472f387266466a4c6956754a50523142563956672b50467976617331654467483245490a77c3f917fdb820cd047682a1398aa14215a1da0697feeec07d06fd29664e3c523f8a1b0aad1de7e4748baef914d19dfcdf3cc22d68e90239b105abd46eaf809450002c75f198ad0b3103e6c78b0daad495fb2f4817a04a7e937d73c16f15ed6d57a2b9d9d842d3"
  ],
  "caveats": [
    {
      "cid": "{\"op\":\"expires\",\"args\":[\"2024-07-16T04:01:58Z\"]}"
    }
  ],
  "sig": [
    "~hex",
    "3b3190a3b479196389e386f770ad80a11e534215cf41a133d04b35b7d8e6f0b9"
  ]
}
2024/07/11 00:00:00 1. Execute Successful Request: /myorg/myapp/do
2024/07/11 00:00:00 Result: map[ok:true]
2024/07/11 00:00:00 2. Execute Failing Request /otherorg/myapp/do
2024/07/11 00:00:00 API Error: do operation failed: 401: macaroon: predicate not satisfied: /macaroon/0xcccdfc9d3c21489abf5dca73d56c88c5/caveat/0: {"op":"org","args":["myorg"]}
2024/07/11 00:00:00 3. Execute Failing Request - Macaroon Verification Failure due to discharge not bound to target
2024/07/11 00:00:00 Target API: macaroon verify failed, debug verification follows
2024/07/11 00:00:00 {
  "traces": [
    {
      "rootKey": "0x92924f5f3cf45e77701bf82253776c0734f4484a4ac98db43ccc858a722572c6",
      "ops": [
        {
          "kind": "HMAC",
          "args": [
            "0x92924f5f3cf45e77701bf82253776c0734f4484a4ac98db43ccc858a722572c6",
            "0xcccdfc9d3c21489abf5dca73d56c88c5"
          ],
          "result": "0x677c8d742425a51d09dbb3037455ad4ef96d5caf71b0694f58764a5e54d0bd84"
        },
        {
          "kind": "HMAC",
          "args": [
            "0x677c8d742425a51d09dbb3037455ad4ef96d5caf71b0694f58764a5e54d0bd84",
            "{\"op\":\"org\",\"args\":[\"myorg\"]}"
          ],
          "result": "0x31f727e0b01c2e41b6a9c8759859702563c424b48f56632854c25894ad962c66"
        },
        {
          "kind": "HMAC",
          "args": [
            "0x31f727e0b01c2e41b6a9c8759859702563c424b48f56632854c25894ad962c66",
            "{\"op\":\"app\",\"args\":[\"myapp\"]}"
          ],
          "result": "0xc0976ce8ef4ed794f941aef859656fc37e8697c2ce731b5ade81fa842ddbbb35"
        },
        {
          "kind": "HMAC",
          "args": [
            "0xc0976ce8ef4ed794f941aef859656fc37e8697c2ce731b5ade81fa842ddbbb35",
            "{\"op\":\"expires\",\"args\":[\"2024-07-16T07:56:58-04:00\"]}"
          ],
          "result": "0x071ffadd5caacc27455b0ee0c57b541b8d3a868feea905c0270c324662a7052a"
        },
        {
          "kind": "Decrypt",
          "args": [
            "0x071ffadd5caacc27455b0ee0c57b541b8d3a868feea905c0270c324662a7052a",
            "0x3c60eba1f59af27b29355651997e79504c9f20f9fc0751d830ee7c27652cdc1f10b1d75f4fa15d508736b822f0516f3303fdaf7020850ba9"
          ],
          "result": "0xf53fe62033bea2654f0d99cd46dc4dd496b97e18fa5ca11ca678396ce809a448"
        },
        {
          "kind": "FAILURE",
          "error": [
            "macaroon: verification failed",
            "macaroon.verify: signatures did not match: want=3b3190a3b479196389e386f770ad80a11e534215cf41a133d04b35b7d8e6f0b9, got=8b3759d3c9c2d046d5b52a5e4a9e39b01944ecc6ea44d3752e5a1093a4bdddcf"
          ]
        }
      ]
    },
    {
      "rootKey": "0xf53fe62033bea2654f0d99cd46dc4dd496b97e18fa5ca11ca678396ce809a448",
      "ops": [
        {
          "kind": "HMAC",
          "args": [
            "0xf53fe62033bea2654f0d99cd46dc4dd496b97e18fa5ca11ca678396ce809a448",
            "0x83a474797065a66167653a7631a36b6964a46b696431a464617461c5010f6167652d656e6372797074696f6e2e6f72672f76310a2d3e205832353531392034374e6c3754476c4a61455357486374524f456d43397351374e474d2b757656744d626d5859693053786b0a78445a573749645a74466d7a632f41562b6a41593253654265306249432b7052496e6c6b597a65444750630a2d2d2d2045324f6e394a5878472f387266466a4c6956754a50523142563956672b50467976617331654467483245490a77c3f917fdb820cd047682a1398aa14215a1da0697feeec07d06fd29664e3c523f8a1b0aad1de7e4748baef914d19dfcdf3cc22d68e90239b105abd46eaf809450002c75f198ad0b3103e6c78b0daad495fb2f4817a04a7e937d73c16f15ed6d57a2b9d9d842d3"
          ],
          "result": "0xac34f4e1d88e140af2c49ccd8aea468caadb34aceddf1f485934e5c8e9de8888"
        },
        {
          "kind": "HMAC",
          "args": [
            "0xac34f4e1d88e140af2c49ccd8aea468caadb34aceddf1f485934e5c8e9de8888",
            "{\"op\":\"expires\",\"args\":[\"2024-07-16T04:01:58Z\"]}"
          ],
          "result": "0x8b3759d3c9c2d046d5b52a5e4a9e39b01944ecc6ea44d3752e5a1093a4bdddcf"
        },
        {
          "kind": "BindForRequest",
          "args": [
            "0xe5cc362e341d84bd24840e90e75ad7c56a409c9afb238e7fd08b1877602d0f20",
            "0x8b3759d3c9c2d046d5b52a5e4a9e39b01944ecc6ea44d3752e5a1093a4bdddcf"
          ],
          "result": "0x3b3190a3b479196389e386f770ad80a11e534215cf41a133d04b35b7d8e6f0b9"
        }
      ]
    }
  ]
}
2024/07/11 00:00:00 API Error: do operation failed: 401: macaroon: verification failed
2024/07/11 00:00:00 Shutting Down API Servers...
2024/07/11 00:00:00 Shutting Down API Servers... complete
```