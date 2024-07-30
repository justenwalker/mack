module bench

go 1.22

toolchain go1.22.5

replace github.com/justenwalker/mack => ../

require (
	github.com/justenwalker/mack v0.0.0
	gopkg.in/macaroon.v2 v2.1.0
)

require (
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	golang.org/x/crypto v0.25.0 // indirect
)

require golang.org/x/sys v0.22.0 // indirect
