package auth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

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
	user, exp, err := parseAccessToken(token)
	if err != nil {
		return r
	}
	if time.Until(exp) < 0 {
		return r
	}
	ctx := WithAuthContext(r.Context(), AuthContext{
		Username: user,
		Time:     time.Now().UTC(),
	})
	r = r.WithContext(ctx)
	return r
}

type accessToken struct {
	Username string `json:"username,omitempty"`
	Expires  string `json:"expires,omitempty"`
}

func createAccessToken(username string) (r LoginResponseBody, err error) {
	expires := 8 * time.Hour
	exp := time.Now().Add(expires).UTC().Format(time.RFC3339)
	bs, err := json.Marshal(accessToken{
		Username: username,
		Expires:  exp,
	})
	if err != nil {
		return r, err
	}
	return LoginResponseBody{
		AccessToken: base64.StdEncoding.EncodeToString(bs),
		ExpiresIn:   int64(expires.Seconds()),
	}, nil
}

func parseAccessToken(str string) (username string, expires time.Time, err error) {
	bs, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", time.Time{}, err
	}
	var at accessToken
	if err = json.Unmarshal(bs, &at); err != nil {
		return "", time.Time{}, err
	}
	t, err := time.Parse(time.RFC3339, at.Expires)
	if err != nil {
		return "", time.Time{}, err
	}
	return at.Username, t, nil
}
