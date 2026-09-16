// Package alias provides unprefixed type aliases for diff command flags.
//
//	import diff "github.com/gloo-foo/cmd-diff/alias"
//	diff.Diff(input, diff.Unified)
package alias

import (
	gloo "github.com/gloo-foo/framework"

	command "github.com/gloo-foo/cmd-diff"
)

// Diff re-exports the constructor by delegation, preserving its exact signature.
func Diff(opts ...any) gloo.Command[[]byte, []byte] { return command.Diff(opts...) }

// DiffInput re-exports the second-input type.
type DiffInput = command.DiffInput

// DiffFs re-exports the filesystem-injection option.
type DiffFs = command.DiffFs

// -u flag: unified diff output
const Unified = command.DiffUnified

// default: ed-style diff output
const NoUnified = command.DiffNoUnified
