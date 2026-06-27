// Package alias provides unprefixed type aliases for diff command flags.
//
//	import diff "github.com/gloo-foo/cmd-diff/alias"
//	diff.Diff(input, diff.Unified)
package alias

import command "github.com/gloo-foo/cmd-diff"

// Diff re-exports the constructor.
var Diff = command.Diff

// DiffInput re-exports the second-input type.
type DiffInput = command.DiffInput

// DiffFs re-exports the filesystem-injection option.
var DiffFs = command.DiffFs

// -u flag: unified diff output
const Unified = command.DiffUnified

// default: ed-style diff output
const NoUnified = command.DiffNoUnified
