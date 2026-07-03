package command

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/destel/rill"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
)

// DiffInput supplies the second input as raw lines, taking precedence over any
// second positional file path.
type DiffInput [][]byte

// lines is one decoded input: the ordered lines diff compares position by
// position.
type lines [][]byte

// Diff compares two line streams position by position and emits their
// differences.
//
// Output format:
//   - default: ed-style — `< line` for an input1-only line, `> line` for an
//     input2-only line, at each position where the lines differ.
//   - DiffUnified (-u): a unified diff with `--- a` / `+++ b` headers, a single
//     `@@ -1,N +1,M @@` hunk, and `-line` / `+line` / ` line` body rows.
//
// Opts:
//   - 1st positional file/Reader: input1 (overrides the upstream stream).
//   - 2nd positional file/Reader: input2.
//   - DiffInput: input2 as raw lines (highest precedence for input2).
//   - DiffUnified: switch output format.
//   - DiffFs: filesystem used to open File positionals (defaults to the OS).
func Diff(opts ...any) gloo.Command[[]byte, []byte] {
	f, rest := foldOptions(opts)
	params := gloo.NewParameters[gloo.File, struct{}](rest...)
	src := newSources(opts, params.Positional, f.fs.value())
	return gloo.FuncCommand[[]byte, []byte](func(ctx context.Context, in gloo.Stream[[]byte]) gloo.Stream[[]byte] {
		return gloo.GenerateFrom(ctx, in, func(_ context.Context, send func([]byte) bool, sendErr func(error)) {
			run(send, sendErr, src, in, f)
		})
	})
}

// run loads both inputs and emits the diff, forwarding any load error.
func run(send func([]byte) bool, sendErr func(error), src sources, in gloo.Stream[[]byte], f flags) {
	input1, input2, err := src.load(in)
	if err != nil {
		sendErr(err)
		return
	}
	formatOf(f).emit(send, input1, input2)
}

// sources resolves the two diff inputs from opts, positionals, and the upstream
// stream. It is an immutable value built once per Diff call.
type sources struct {
	fs             afero.Fs
	positionals    []any
	explicitInput2 lines
	hasExplicit2   bool
}

// newSources classifies the opts into the resolved input sources.
func newSources(opts, positionals []any, fs afero.Fs) sources {
	explicit, ok := explicitInput2(opts)
	return sources{
		fs:             fs,
		positionals:    positionals,
		explicitInput2: explicit,
		hasExplicit2:   ok,
	}
}

// explicitInput2 returns the first DiffInput option, if any.
func explicitInput2(opts []any) (lines, bool) {
	for _, o := range opts {
		if v, ok := o.(DiffInput); ok {
			return lines(v), true
		}
	}
	return nil, false
}

// load resolves input1 then input2.
func (s sources) load(in gloo.Stream[[]byte]) (lines, lines, error) {
	input1, err := s.loadInput1(in)
	if err != nil {
		return nil, nil, err
	}
	input2, err := s.loadInput2()
	if err != nil {
		return nil, nil, err
	}
	return input1, input2, nil
}

// loadInput1 reads the first positional, falling back to the upstream stream.
func (s sources) loadInput1(in gloo.Stream[[]byte]) (lines, error) {
	if len(s.positionals) >= 1 {
		return s.readPositional(s.positionals[0])
	}
	got, err := rill.ToSlice(in.Chan())
	return lines(got), err
}

// loadInput2 prefers an explicit DiffInput, else the second positional, else
// nothing.
func (s sources) loadInput2() (lines, error) {
	switch {
	case s.hasExplicit2:
		return s.explicitInput2, nil
	case len(s.positionals) >= 2:
		return s.readPositional(s.positionals[1])
	default:
		return nil, nil
	}
}

// readPositional decodes one positional argument into lines. The framework
// guarantees every positional is a gloo.File path or an io.Reader (see
// gloo.NewParameters), so those two cases are exhaustive.
func (s sources) readPositional(positional any) (lines, error) {
	if name, ok := positional.(gloo.File); ok {
		return s.readFile(name)
	}
	return scanLines(positional.(io.Reader))
}

// readFile opens a File positional on the injected filesystem and scans it.
func (s sources) readFile(name gloo.File) (lines, error) {
	f, err := s.fs.Open(string(name))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return scanLines(f)
}

// scanLines reads r into a slice of independently-owned line copies.
func scanLines(r io.Reader) (lines, error) {
	scanner := bufio.NewScanner(r)
	var out lines
	for scanner.Scan() {
		out = append(out, bytes.Clone(scanner.Bytes()))
	}
	return out, scanner.Err()
}

// formatter renders the diff of two inputs through send.
type formatter func(send func([]byte) bool, input1, input2 lines)

// emit applies the formatter.
func (fn formatter) emit(send func([]byte) bool, input1, input2 lines) {
	fn(send, input1, input2)
}

// formatOf selects the output formatter for the configured flags.
func formatOf(f flags) formatter {
	if bool(f.unifiedEnabled) {
		return emitUnified
	}
	return emitSimple
}

// longer returns the greater of the two input lengths.
func longer(input1, input2 lines) int {
	if len(input2) > len(input1) {
		return len(input2)
	}
	return len(input1)
}

// at returns the line at index i, or nil when i is past the end.
func (l lines) at(i int) []byte {
	if i < len(l) {
		return l[i]
	}
	return nil
}

// emitSimple renders the default ed-style diff: each differing position
// contributes a `< ` row for its input1 line and a `> ` row for its input2 line.
func emitSimple(send func([]byte) bool, input1, input2 lines) {
	for i := range longer(input1, input2) {
		emitSimpleAt(send, input1, input2, position(i))
	}
}

// position is an index into the paired inputs: line N of input1 aligned with
// line N of input2.
type position int

// emitSimpleAt renders one position of the ed-style diff.
func emitSimpleAt(send func([]byte) bool, input1, input2 lines, i position) {
	if bytes.Equal(input1.at(int(i)), input2.at(int(i))) {
		return
	}
	if int(i) < len(input1) {
		send(fmt.Appendf(nil, "< %s", input1[i]))
	}
	if int(i) < len(input2) {
		send(fmt.Appendf(nil, "> %s", input2[i]))
	}
}

// emitUnified renders the unified-diff format, including headers and a single
// hunk, when the inputs differ; identical inputs produce no output.
func emitUnified(send func([]byte) bool, input1, input2 lines) {
	if input1.equal(input2) {
		return
	}
	send([]byte("--- a"))
	send([]byte("+++ b"))
	send(fmt.Appendf(nil, "@@ -1,%d +1,%d @@", len(input1), len(input2)))
	for i := range longer(input1, input2) {
		emitUnifiedAt(send, input1, input2, position(i))
	}
}

// emitUnifiedAt renders one position of the unified-diff body.
func emitUnifiedAt(send func([]byte) bool, input1, input2 lines, i position) {
	switch {
	case int(i) >= len(input1):
		send(fmt.Appendf(nil, "+%s", input2[i]))
	case int(i) >= len(input2):
		send(fmt.Appendf(nil, "-%s", input1[i]))
	case !bytes.Equal(input1[i], input2[i]):
		send(fmt.Appendf(nil, "-%s", input1[i]))
		send(fmt.Appendf(nil, "+%s", input2[i]))
	default:
		send(fmt.Appendf(nil, " %s", input1[i]))
	}
}

// equal reports whether two inputs are identical line for line.
func (l lines) equal(other lines) bool {
	if len(l) != len(other) {
		return false
	}
	for i := range l {
		if !bytes.Equal(l[i], other[i]) {
			return false
		}
	}
	return true
}
