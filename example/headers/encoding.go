package headers

import (
	"encoding/base64"
	"fmt"

	"github.com/justenwalker/mack/encoding/msgpack"
	"github.com/justenwalker/mack/macaroon"
)

const MacaroonBearerPrefix = "macaroon.v1."

func EncodeMacaroonStack(stack macaroon.Stack) (string, error) {
	enc, err := msgpack.Encoding.EncodeStack(stack)
	if err != nil {
		return "", err
	}
	return MacaroonBearerPrefix + base64.StdEncoding.EncodeToString(enc), nil
}

func DecodeMacaroonStack(token string) (macaroon.Stack, error) {
	stackBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return macaroon.Stack{}, fmt.Errorf("decode token failed: %w", err)
	}
	var stack macaroon.Stack
	err = msgpack.Encoding.DecodeStack(stackBytes, &stack)
	if err != nil {
		return macaroon.Stack{}, fmt.Errorf("decode token failed: %w", err)
	}
	return stack, nil
}
