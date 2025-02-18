package auth

import (
	"context"
	"time"
)

type Authorization struct {
	Username string
	Time     time.Time
}

type authnContextKey struct{}

func WithAuthorization(ctx context.Context, rc Authorization) context.Context {
	return context.WithValue(ctx, authnContextKey{}, rc)
}

func AuthorizationFromContext(ctx context.Context) (Authorization, bool) {
	v := ctx.Value(authnContextKey{})
	if v == nil {
		return Authorization{}, false
	}
	rc, ok := v.(Authorization)
	return rc, ok
}
