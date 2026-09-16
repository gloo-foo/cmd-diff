package alias_test

import (
	"slices"
	"testing"

	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/testable"
	"github.com/spf13/afero"

	diff "github.com/gloo-foo/cmd-diff/alias"
)

// The alias package re-exports the constructor and flag constants under
// unprefixed names. A mis-wired re-export (say, Unified bound to the disabled
// constant, or Diff bound to the wrong function) compiles cleanly, so only
// behavior can prove the wiring. Each test exercises one re-export and asserts
// the exact diff output it must produce.
//
// input1 = "a b c" (from the stream); input2 = "a x c" (via the re-exported
// DiffInput), so line 2 differs and lines 1 and 3 match.

const diffInput1 = "a\nb\nc\n"

func input2() diff.DiffInput {
	return diff.DiffInput{[]byte("a"), []byte("x"), []byte("c")}
}

func assertLines(t *testing.T, got, want []string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAlias_DefaultEdStyle(t *testing.T) {
	// Default (no Unified): ed-style `< `/`> ` rows for the single differing line.
	lines, err := testable.TestLines(diff.Diff(input2()), diffInput1)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"< b", "> x"})
}

func TestAlias_Unified(t *testing.T) {
	// Unified (-u): headers, one hunk, then -/+/' ' body rows.
	lines, err := testable.TestLines(diff.Diff(input2(), diff.Unified), diffInput1)
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

func TestAlias_NoUnifiedMatchesDefault(t *testing.T) {
	// NoUnified is the disabled form: it must behave like passing no flag at all.
	lines, err := testable.TestLines(diff.Diff(input2(), diff.NoUnified), diffInput1)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"< b", "> x"})
}

func TestAlias_Fs(t *testing.T) {
	// DiffFs injects the filesystem used to open File positionals.
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "1.txt", []byte("a\nb\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := afero.WriteFile(fs, "2.txt", []byte("a\nx\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lines, err := testable.TestLines(
		diff.Diff(gloo.File("1.txt"), gloo.File("2.txt"), diff.DiffFs{Fs: fs}),
		"",
	)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"< b", "> x"})
}
