package headers

import (
	"net/http"
	"strings"

	"github.com/justenwalker/mack/macaroon"
)

func SetMacaroonStackAuthorization(header http.Header, stack macaroon.Stack) {
	if enc, err := EncodeMacaroonStack(stack); err == nil {
		header.Set("Authorization", "Bearer "+enc)
	}
}

func GetBearerToken(header http.Header) (string, bool) {
	auth := header.Get("Authorization")
	if len(auth) < 7 {
		return "", false
	}
	if !strings.EqualFold(auth[:7], "bearer ") {
		return "", false
	}
	return auth[7:], true
}
