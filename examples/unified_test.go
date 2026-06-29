package diff_test

import (
	gloo "github.com/gloo-foo/framework/patterns"

	command "github.com/gloo-foo/cmd-diff"
)

func ExampleDiff_unified() {
	// diff -u file1.txt file2.txt — unified format.
	gloo.MustRun(
		command.Diff("testdata/file1.txt", "testdata/file2.txt", command.DiffUnified),
	)
	// Output:
	// --- a
	// +++ b
	// @@ -1,3 +1,3 @@
	//  alpha
	// -beta
	// +BETA
	//  gamma
}
