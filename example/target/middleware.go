package target

import (
	"net/http"
	"strings"

	"example/headers"
)

func AuthorizeMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, setAuthContext(r))
	})
}

func setAuthContext(r *http.Request) *http.Request {
	token, ok := headers.GetBearerToken(r.Header)
	if !ok {
		return r
	}
	if strings.HasPrefix(token, headers.MacaroonBearerPrefix) {
		if stack, err := headers.DecodeMacaroonStack(token[len(headers.MacaroonBearerPrefix):]); err == nil {
			return r.WithContext(WithAuthContext(r.Context(), AuthContext{
				Token: token,
				Stack: &stack,
			}))
		}
	}
	return r.WithContext(WithAuthContext(r.Context(), AuthContext{
		Token: token,
	}))
}
