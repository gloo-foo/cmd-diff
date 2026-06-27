package diff_test

import (
	. "github.com/gloo-foo/cmd-diff"
	gloo "github.com/gloo-foo/framework/patterns"
)

func ExampleDiff_unified() {
	// diff -u file1.txt file2.txt — unified format.
	gloo.MustRun(
		Diff("testdata/file1.txt", "testdata/file2.txt", DiffUnified),
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
