//go:build tools

package mack

import (
	_ "github.com/daixiang0/gci"
	_ "go.uber.org/mock/mockgen"
	_ "golang.org/x/tools/cmd/stringer"
	_ "mvdan.cc/gofumpt"
)
