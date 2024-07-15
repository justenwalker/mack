package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"filippo.io/age"

	"github.com/justenwalker/mack/crypt/agecrypt"
	"github.com/justenwalker/mack/encoding/msgpack"
	"github.com/justenwalker/mack/exchange"
	"github.com/justenwalker/mack/macaroon"
	"github.com/justenwalker/mack/macaroon/thirdparty"
)

//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=oapi-codegen.config.yaml openapi.yaml

type API struct {
	location   string
	scheme     *macaroon.Scheme
	recipient  agecrypt.Recipient
	discharger *thirdparty.Discharger
}

func (as *API) PostValidateToken(w http.ResponseWriter, r *http.Request) {
	var validate ValidateTokenRequest
	if err := readModel(r, &validate); err != nil {
		as.writeError(w, http.StatusBadRequest, err)
		return
	}
	user, exp, err := parseAccessToken(validate.AccessToken)
	if err != nil {
		as.writeError(w, http.StatusBadRequest, err)
		return
	}
	writeModel(w, http.StatusOK, ValidateTokenResponse{
		Expires:  exp,
		Username: user,
	})
}

func NewAPI(scheme *macaroon.Scheme, location string) (*API, error) {
	thirdPartyID, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("age.GenerateX25519Identity: %w", err)
	}
	discharger, err := createDischarger(scheme, location, []agecrypt.Identity{
		{
			KeyID:     "kid1",
			Identity:  thirdPartyID,
			Recipient: thirdPartyID.Recipient(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("CreateDischarger: %w", err)
	}
	return &API{
		scheme:   scheme,
		location: location,
		recipient: agecrypt.Recipient{
			KeyID:     "kid1",
			Recipient: thirdPartyID.Recipient(),
		},
		discharger: discharger,
	}, nil
}

func (as *API) Handler() http.Handler {
	return HandlerWithOptions(as, StdHTTPServerOptions{
		Middlewares: []MiddlewareFunc{
			AuthorizeMiddleware,
		},
	})
}

func (as *API) PostDischarge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ac, ok := AuthFromContext(ctx)
	if !ok {
		as.writeError(w, http.StatusForbidden, errors.New("forbidden"))
		return
	}
	var dmr DischargeMacaroonRequest
	if err := readModel(r, &dmr); err != nil {
		as.writeError(w, http.StatusBadRequest, err)
		return
	}
	cavID, err := base64.StdEncoding.DecodeString(dmr.CaveatId)
	if err != nil {
		as.writeError(w, http.StatusBadRequest, err)
		return
	}
	m, err := as.discharger.Discharge(ctx, cavID, PredicateChecker{
		AuthContext: ac,
	})
	if err != nil {
		as.writeError(w, http.StatusBadRequest, err)
		return
	}
	expiresIn := 5 * time.Minute
	m, err = as.scheme.AddFirstPartyCaveat(&m, caveatBytes("expires", time.Now().UTC().Add(expiresIn).Format(time.RFC3339)))
	if err != nil {
		as.writeError(w, http.StatusBadRequest, err)
		return
	}
	bs, err := msgpack.Encoding.EncodeMacaroon(&m)
	if err != nil {
		as.writeError(w, http.StatusBadRequest, err)
		return
	}
	writeModel(w, http.StatusOK, DischargeMacaroonResponseBody{
		ExpiresIn: int64(expiresIn.Seconds()),
		Macaroon:  base64.StdEncoding.EncodeToString(bs),
	})
}

func (as *API) GetIdentities(w http.ResponseWriter, r *http.Request) {
	pubKey := fmt.Sprintf("%v", as.recipient.Recipient)
	writeModel(w, http.StatusOK, IdentitiesResponseBody{
		{
			KeyId:     as.recipient.KeyID,
			KeyType:   "age",
			PublicKey: base64.StdEncoding.EncodeToString([]byte(pubKey)),
		},
	})
}

func (as *API) PostLogin(w http.ResponseWriter, r *http.Request) {
	var login LoginRequestBody
	if err := readModel(r, &login); err != nil {
		as.writeError(w, http.StatusBadRequest, err)
		return
	}
	if login.Password != "secret" {
		as.writeError(w, http.StatusForbidden, errors.New("bad password"))
		return
	}
	resp, err := createAccessToken(login.Username)
	if err != nil {
		as.writeError(w, http.StatusInternalServerError, err)
	}
	writeModel(w, http.StatusOK, resp)
}

var _ ServerInterface = (*API)(nil)

func createDischarger(scheme *macaroon.Scheme, location string, ids []agecrypt.Identity) (*thirdparty.Discharger, error) {
	dec := agecrypt.NewDecryptor(ids)
	discharger, err := thirdparty.NewDischarger(thirdparty.DischargerConfig{
		Location: location,
		Scheme:   scheme,
		TicketExtractor: &exchange.TicketExtractor{
			Decryptor: dec,
			Decoder:   msgpack.Encoding,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("thirdparty.NewDischarger: %w", err)
	}
	return discharger, nil
}

func readModel[T any](r *http.Request, b *T) error {
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, b)
}

func writeModel[T any](w http.ResponseWriter, code int, t T) {
	out, _ := json.Marshal(t)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(out)
}

func (as *API) writeError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	bs, _ := json.Marshal(ErrorResponseBody{
		Code:  code,
		Error: err.Error(),
	})
	_, _ = w.Write(bs)
}
