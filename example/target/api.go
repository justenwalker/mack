package target

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"filippo.io/age"
	"github.com/google/uuid"

	"example/agecrypt"
	"example/auth"
	"example/msgpack"

	"github.com/justenwalker/mack"
	"github.com/justenwalker/mack/thirdparty"
	"github.com/justenwalker/mack/thirdparty/exchange"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=oapi-codegen.config.yaml openapi.yaml

type API struct {
	secretKey []byte
	location  string
	scheme    *mack.Scheme
	tps       *thirdparty.Attenuator
	authSvc   *auth.ClientWithResponses
}

type APIConfig struct {
	Scheme      *mack.Scheme
	Location    string
	SecretKey   string
	AuthService string
}

func NewAPI(ctx context.Context, cfg APIConfig) (*API, error) {
	client, err := auth.NewClientWithResponses(cfg.AuthService)
	if err != nil {
		return nil, err
	}
	recp, err := getAuthServiceRecipient(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("get auth service public key: %w", err)
	}
	enc, err := agecrypt.NewEncryptor(recp)
	if err != nil {
		return nil, err
	}
	tps, err := thirdparty.NewAttenuator(thirdparty.AttenuatorConfig{
		Location: cfg.Location,
		Scheme:   cfg.Scheme,
		CaveatIssuer: &exchange.CaveatIDIssuer{
			Encryptor: enc,
			Encoder:   msgpack.Encoding,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("CreateThirdPartyService: %w", err)
	}
	return &API{
		scheme:    cfg.Scheme,
		secretKey: []byte(cfg.SecretKey),
		location:  cfg.Location,
		tps:       tps,
		authSvc:   client,
	}, nil
}

func (as *API) GetMacaroonRequest(w http.ResponseWriter, r *http.Request, params GetMacaroonRequestParams) {
	if params.Org == nil {
		as.writeError(w, http.StatusBadRequest, errors.New("must provide an org"))
	}
	caveats := [][]byte{
		caveatBytes("org", *params.Org),
	}
	if params.App != nil {
		caveats = append(caveats, caveatBytes("app", *params.App))
	}
	ac, ok := AuthFromContext(r.Context())
	if !ok {
		as.writeError(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	resp, err := as.authSvc.PostValidateTokenWithResponse(r.Context(), auth.PostValidateTokenJSONRequestBody{
		AccessToken: ac.Token,
	})
	if err != nil {
		as.writeError(w, http.StatusInternalServerError, err)
	}
	if resp.JSON200 == nil {
		as.writeError(w, http.StatusForbidden, err)
		return
	}
	username := resp.JSON200.Username
	id, _ := uuid.NewRandom()
	key, _ := as.keyID(id[:])
	exp := 8 * time.Hour
	m, err := as.scheme.NewMacaroon(as.location, id[:], key, append(caveats, caveatBytes("expires", time.Now().Add(exp).Format(time.RFC3339)))...)
	if err != nil {
		as.writeError(w, http.StatusInternalServerError, err)
		return
	}
	m, err = as.tps.Attenuate(r.Context(), &m, caveatBytes("user", username))
	if err != nil {
		as.writeError(w, http.StatusInternalServerError, err)
		return
	}
	mp, err := msgpack.Encoding.EncodeMacaroon(&m)
	if err != nil {
		as.writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeModel(w, http.StatusOK, MacaroonResponse{
		ExpiresIn: int64(exp.Seconds()),
		Macaroon:  base64.StdEncoding.EncodeToString(mp),
	})
}

func (as *API) PostOrgAppDo(w http.ResponseWriter, r *http.Request, org string, app string) {
	ac, ok := AuthFromContext(r.Context())
	if !ok || ac.Stack == nil {
		as.writeError(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	stack := *ac.Stack
	key, err := as.keyID(stack.Target().ID())
	if err != nil {
		as.writeError(w, http.StatusUnauthorized, err)
		return
	}
	ctx := mack.WithVerifyContext(r.Context())
	vc, err := as.scheme.Verify(ctx, key, stack)
	if err != nil {
		traces := mack.GetTraces(ctx)
		log.Println("Target API: macaroon verify failed, debug verification follows")
		log.Println(traces.String())
		as.writeError(w, http.StatusUnauthorized, err)
		return
	}
	err = vc.Clear(r.Context(), PredicateChecker{
		RequestContext: RequestContext{
			Org:  org,
			App:  app,
			Time: time.Now(),
		},
	})
	if err != nil {
		as.writeError(w, http.StatusUnauthorized, err)
		return
	}
	writeModel(w, http.StatusOK, OperationResponse{
		"ok": true,
	})
}

func (as *API) Handler() http.Handler {
	return HandlerWithOptions(as, StdHTTPServerOptions{
		Middlewares: []MiddlewareFunc{
			AuthorizeMiddleware,
		},
	})
}

func getAuthServiceRecipient(ctx context.Context, client auth.ClientWithResponsesInterface) (agecrypt.Recipient, error) {
	resp, err := client.GetIdentitiesWithResponse(ctx)
	if err != nil {
		return agecrypt.Recipient{}, err
	}
	ids := *resp.JSONDefault
	if len(ids) == 0 {
		return agecrypt.Recipient{}, errors.New("unable to find recipient")
	}
	pubkey, err := base64.StdEncoding.DecodeString(ids[0].PublicKey)
	if err != nil {
		return agecrypt.Recipient{}, err
	}
	recps, err := age.ParseRecipients(bytes.NewReader(pubkey))
	if err != nil {
		return agecrypt.Recipient{}, err
	}
	return agecrypt.Recipient{
		KeyID:     ids[0].KeyId,
		Recipient: recps[0],
	}, nil
}

var _ ServerInterface = (*API)(nil)

func (as *API) keyID(id []byte) ([]byte, error) {
	return GenerateKeyArgon2ID(as.secretKey, id, as.scheme.KeySize())
}

func writeModel[T any](w http.ResponseWriter, code int, t T) {
	out, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(out)
}

func (as *API) writeError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	bs, err := json.Marshal(ErrorResponseBody{
		Code:  code,
		Error: err.Error(),
	})
	if err != nil {
		panic(err)
	}
	_, _ = w.Write(bs)
}
