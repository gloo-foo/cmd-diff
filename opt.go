package command

import "github.com/spf13/afero"

// diffUnifiedFlag enables unified diff output (-u). Default off.
type diffUnifiedFlag bool

const (
	DiffUnified   diffUnifiedFlag = true
	DiffNoUnified diffUnifiedFlag = false
)

func (u diffUnifiedFlag) Configure(flags *flags) { flags.unified = u }

// diffFs injects the filesystem used to open File positionals, so tests can
// supply an in-memory filesystem. The zero value falls back to the OS.
type diffFs struct{ afero.Fs }

// DiffFs selects the filesystem diff uses to open File positional arguments.
func DiffFs(fs afero.Fs) diffFs { return diffFs{fs} }

func (f diffFs) Configure(flags *flags) { flags.fs = f }

// value returns the configured filesystem, defaulting to the OS filesystem.
func (f diffFs) value() afero.Fs {
	if f.Fs == nil {
		return afero.NewOsFs()
	}
	return f.Fs
}

type flags struct {
	unified diffUnifiedFlag
	fs      diffFs
}
