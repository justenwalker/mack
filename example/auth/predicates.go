package auth

import (
	"context"
	"encoding/json"
	"fmt"
)

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

type Caveat struct {
	Op   string   `json:"op"`
	Args []string `json:"args"`
}

type PredicateChecker struct {
	AuthContext AuthContext
}

func (p PredicateChecker) CheckPredicate(ctx context.Context, predicate []byte) (bool, error) {
	c, err := caveatFromBytes(predicate)
	if err != nil {
		return false, fmt.Errorf("failed to decode caveat '%s': %w", string(predicate), err)
	}
	switch c.Op {
	case "user":
		if len(c.Args) == 0 {
			return false, nil
		}
		return p.AuthContext.Username == c.Args[0], nil
	default:
		return false, fmt.Errorf("unexpected caveat %s: %v", c.Op, c.Args)
	}
}
