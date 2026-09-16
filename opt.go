package command

import "github.com/spf13/afero"

// diffUnifiedFlag enables unified diff output (-u). Default off.
type diffUnifiedFlag bool

const (
	DiffUnified   diffUnifiedFlag = true
	DiffNoUnified diffUnifiedFlag = false
)

// DiffFs injects the filesystem used to open File positionals, so tests can
// supply an in-memory filesystem. The zero value falls back to the OS.
type DiffFs struct{ afero.Fs }

// value returns the configured filesystem, defaulting to the OS filesystem.
func (f DiffFs) value() afero.Fs {
	if f.Fs == nil {
		return afero.NewOsFs()
	}
	return f.Fs
}

// flags holds the parsed diff options.
type flags struct {
	fs             DiffFs
	unifiedEnabled diffUnifiedFlag
}

// with folds one option value into the flags, returning the updated copy and
// whether the argument was one of this command's option types (false leaves it
// for positional classification).
func (f flags) with(o any) (flags, bool) {
	switch v := o.(type) {
	case diffUnifiedFlag:
		f.unifiedEnabled = v
	case DiffFs:
		f.fs = v
	default:
		return f, false
	}
	return f, true
}

// foldOptions folds the command's own option values into a flags value and
// returns the remaining arguments (positional inputs) in their original order.
func foldOptions(opts []any) (flags, []any) {
	var f flags
	var rest []any
	for _, o := range opts {
		next, isOption := f.with(o)
		if !isOption {
			rest = append(rest, o)
			continue
		}
		f = next
	}
	return f, rest
}
