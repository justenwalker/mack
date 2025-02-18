package mack

//go:generate go tool -modfile tools.mod github.com/matryer/moq -pkg mack_test -out ./mack.moq_test.go . HMACScheme EncryptionScheme BindForRequestScheme PredicateChecker
//go:generate go tool -modfile tools.mod github.com/matryer/moq -pkg thirdparty_test -out ./thirdparty/thirdparty.moq_test.go ./thirdparty ThirdParty TicketExtractor CaveatIDIssuer PredicateChecker
//go:generate go tool -modfile tools.mod github.com/matryer/moq -pkg exchange_test -out ./thirdparty/exchange/exchange.moq_test.go ./thirdparty/exchange Encoder Encryptor Decoder Decryptor
