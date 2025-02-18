package headers

import (
	"encoding/base64"
	"fmt"

	"example/msgpack"

	"github.com/justenwalker/mack"
)

const MacaroonBearerPrefix = "mack.v1." //nolint:gosec // not a credential

func EncodeMacaroonStack(stack mack.Stack) (string, error) {
	enc, err := msgpack.Encoding.EncodeStack(stack)
	if err != nil {
		return "", err
	}
	return MacaroonBearerPrefix + base64.StdEncoding.EncodeToString(enc), nil
}

func DecodeMacaroonStack(token string) (mack.Stack, error) {
	stackBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return mack.Stack{}, fmt.Errorf("decode token failed: %w", err)
	}
	var stack mack.Stack
	err = msgpack.Encoding.DecodeStack(stackBytes, &stack)
	if err != nil {
		return mack.Stack{}, fmt.Errorf("decode token failed: %w", err)
	}
	return stack, nil
}
