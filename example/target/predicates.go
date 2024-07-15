package target

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/justenwalker/mack/macaroon/thirdparty"
)

type RequestContext struct {
	Org  string
	App  string
	Time time.Time
}

var _ thirdparty.PredicateChecker = PredicateChecker{}

type PredicateChecker struct {
	RequestContext RequestContext
}

func (p PredicateChecker) CheckPredicate(ctx context.Context, predicate []byte) (bool, error) {
	c, err := caveatFromBytes(predicate)
	if err != nil {
		return false, fmt.Errorf("failed to decode caveat '%s': %w", string(predicate), err)
	}
	switch c.Op {
	case "org":
		if len(c.Args) == 0 {
			return false, nil
		}
		return p.RequestContext.Org == c.Args[0], nil
	case "app":
		if len(c.Args) == 0 {
			return false, nil
		}
		return p.RequestContext.App == c.Args[0], nil
	case "expires":
		if len(c.Args) == 0 {
			return false, nil
		}
		var exp time.Time
		exp, err = time.Parse(time.RFC3339, c.Args[0])
		if err != nil {
			return false, fmt.Errorf("could not parse time '%s': %w", c.Args[0], err)
		}
		return p.RequestContext.Time.Before(exp), nil
	default:
		return false, fmt.Errorf("unexpected caveat %s: %v", c.Op, c.Args)
	}
}

type Caveat struct {
	Op   string   `json:"op"`
	Args []string `json:"args"`
}

func caveatBytes(op string, args ...string) []byte {
	bs, _ := json.Marshal(Caveat{Op: op, Args: args})
	return bs
}

func caveatFromBytes(bs []byte) (Caveat, error) {
	var c Caveat
	err := json.Unmarshal(bs, &c)
	if err != nil {
		return c, fmt.Errorf("json.Unmarshal: %w", err)
	}
	return c, nil
}
