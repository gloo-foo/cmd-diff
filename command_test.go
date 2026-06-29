package command_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/testable"
	"github.com/spf13/afero"

	command "github.com/gloo-foo/cmd-diff"
)

// diff compares two inputs position by position (line N of input1 vs line N of
// input2). The default output is ed-style:
//   - `< line`  for an input1-only line at a differing position
//   - `> line`  for an input2-only line at a differing position
//
// The -u (DiffUnified) output is a unified diff: `--- a` / `+++ b` headers, a
// single `@@ -1,N +1,M @@` hunk, and `-`/`+`/' ' body rows. These tests assert
// the exact bytes diff emits, including the no-output case for identical inputs.

func input2(ss ...string) command.DiffInput {
	in := make(command.DiffInput, len(ss))
	for i, s := range ss {
		in[i] = []byte(s)
	}
	return in
}

func assertLines(t *testing.T, got, want []string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDiff_IdenticalProducesNoOutput(t *testing.T) {
	lines, err := testable.TestLines(command.Diff(input2("a", "b", "c")), "a\nb\nc\n")
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{})
}

func TestDiff_ChangedLine(t *testing.T) {
	// Line 2 differs; lines 1 and 3 match.
	lines, err := testable.TestLines(command.Diff(input2("a", "x", "c")), "a\nb\nc\n")
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"< b", "> x"})
}

func TestDiff_DeletedTrailingLine(t *testing.T) {
	// input1 is longer: its trailing line has no counterpart in input2.
	lines, err := testable.TestLines(command.Diff(input2("a")), "a\nb\n")
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"< b"})
}

func TestDiff_AddedTrailingLine(t *testing.T) {
	// input2 is longer: its trailing line has no counterpart in input1.
	lines, err := testable.TestLines(command.Diff(input2("a", "b")), "a\n")
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"> b"})
}

func TestDiff_UnifiedIdenticalProducesNoOutput(t *testing.T) {
	lines, err := testable.TestLines(
		command.Diff(input2("a", "b"), command.DiffUnified),
		"a\nb\n",
	)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{})
}

func TestDiff_UnifiedChangedLine(t *testing.T) {
	lines, err := testable.TestLines(
		command.Diff(input2("a", "x", "c"), command.DiffUnified),
		"a\nb\nc\n",
	)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{
		"--- a",
		"+++ b",
		"@@ -1,3 +1,3 @@",
		" a",
		"-b",
		"+x",
		" c",
	})
}

func TestDiff_UnifiedDeletedTrailingLine(t *testing.T) {
	lines, err := testable.TestLines(
		command.Diff(input2("a"), command.DiffUnified),
		"a\nb\n",
	)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{
		"--- a",
		"+++ b",
		"@@ -1,2 +1,1 @@",
		" a",
		"-b",
	})
}

func TestDiff_UnifiedAddedTrailingLine(t *testing.T) {
	lines, err := testable.TestLines(
		command.Diff(input2("a", "b"), command.DiffUnified),
		"a\n",
	)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{
		"--- a",
		"+++ b",
		"@@ -1,1 +1,2 @@",
		" a",
		"+b",
	})
}

func TestDiff_NoUnifiedMatchesDefault(t *testing.T) {
	// The DiffNoUnified constant is the disabled form: it must behave like no flag.
	lines, err := testable.TestLines(
		command.Diff(input2("a", "x"), command.DiffNoUnified),
		"a\nb\n",
	)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"< b", "> x"})
}

func TestDiff_EmptyInputs(t *testing.T) {
	lines, err := testable.TestLines(command.Diff(command.DiffInput{}), "")
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{})
}

func TestDiff_NoSecondInput(t *testing.T) {
	// With neither a DiffInput nor a second positional, input2 is empty: every
	// input1 line differs and emits a `< ` row.
	lines, err := testable.TestLines(
		command.Diff(strings.NewReader("a\nb\n")),
		"",
	)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"< a", "< b"})
}

func TestDiff_PositionalFilesViaMemFs(t *testing.T) {
	// Both inputs as File positionals, read through an injected filesystem.
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "a.txt", []byte("a\nb\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := afero.WriteFile(fs, "b.txt", []byte("a\nx\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lines, err := testable.TestLines(
		command.Diff(gloo.File("a.txt"), gloo.File("b.txt"), command.DiffFs(fs)),
		"ignored upstream\n",
	)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"< b", "> x"})
}

func TestDiff_ReaderPositionals(t *testing.T) {
	// Both inputs as io.Reader positionals.
	lines, err := testable.TestLines(
		command.Diff(strings.NewReader("a\nb\n"), strings.NewReader("a\nx\n")),
		"",
	)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"< b", "> x"})
}

func TestDiff_FileNotFoundPropagates(t *testing.T) {
	fs := afero.NewMemMapFs()
	_, err := testable.TestLines(
		command.Diff(gloo.File("missing.txt"), command.DiffFs(fs)),
		"",
	)
	if err == nil {
		t.Fatal("expected an error opening a missing file, got nil")
	}
}

func TestDiff_Input2FileNotFoundPropagates(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "a.txt", []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := testable.TestLines(
		command.Diff(gloo.File("a.txt"), gloo.File("missing.txt"), command.DiffFs(fs)),
		"",
	)
	if err == nil {
		t.Fatal("expected an error opening the missing second file, got nil")
	}
}

func TestDiff_ScannerErrorPropagates(t *testing.T) {
	// A reader that fails mid-scan must surface its error, not be swallowed.
	_, err := testable.TestLines(
		command.Diff(errReader{}),
		"",
	)
	if !errors.Is(err, errBoom) {
		t.Fatalf("got %v, want %v", err, errBoom)
	}
}

func TestDiff_DefaultFsResolves(t *testing.T) {
	// With no DiffFs option, diff opens File positionals on the OS filesystem.
	lines, err := testable.TestLines(
		command.Diff(gloo.File("testdata/a.txt"), gloo.File("testdata/b.txt")),
		"",
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) == 0 {
		t.Fatal("expected output from on-disk testdata files")
	}
}

var errBoom = errors.New("boom")

// errReader is an io.Reader that always fails, exercising the scanner error path.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errBoom }
