package diff_test

import (
	gloo "github.com/gloo-foo/framework/patterns"

	command "github.com/gloo-foo/cmd-diff"
)

func ExampleDiff_basic() {
	// diff file1.txt file2.txt — line 2 differs.
	if err := gloo.Run(
		command.Diff("testdata/file1.txt", "testdata/file2.txt"),
	); err != nil {
		panic(err)
	}
	// Output:
	// < beta
	// > BETA
}
