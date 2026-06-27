package diff_test

import (
	. "github.com/gloo-foo/cmd-diff"
	gloo "github.com/gloo-foo/framework/patterns"
)

func ExampleDiff_basic() {
	// diff file1.txt file2.txt — line 2 differs.
	gloo.MustRun(
		Diff("testdata/file1.txt", "testdata/file2.txt"),
	)
	// Output:
	// < beta
	// > BETA
}
