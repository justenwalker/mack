package target

import (
	"context"

	"github.com/justenwalker/mack/macaroon"
)

type AuthContext struct {
	Token string
	Stack *macaroon.Stack
}

type authContextKey struct{}

func WithAuthContext(ctx context.Context, rc AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey{}, rc)
}

func AuthFromContext(ctx context.Context) (AuthContext, bool) {
	v := ctx.Value(authContextKey{})
	if v == nil {
		return AuthContext{}, false
	}
	rc, ok := v.(AuthContext)
	return rc, ok
}
