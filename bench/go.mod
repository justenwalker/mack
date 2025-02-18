module bench

go 1.24

toolchain go1.24.0

replace github.com/justenwalker/mack => ../

require (
	github.com/justenwalker/mack v0.0.0
	gopkg.in/macaroon.v2 v2.1.0
)

require (
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	golang.org/x/crypto v0.33.0 // indirect
)

require golang.org/x/sys v0.30.0 // indirect
