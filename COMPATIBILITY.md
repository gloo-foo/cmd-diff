# Diff Command Compatibility

## Summary

Largely divergent from GNU `diff`: `diff` compares two files **positionally** (line N vs line N) rather than computing an LCS edit script, and its output format (ed-style `<`/`>` by default, single-hunk unified under `-u`) does not match any GNU `diff` mode. Only the identical-files case (no output, exit 0) is byte-identical to GNU.

## Key Behaviors

```bash
# Identical files: no output, exit 0 (matches GNU diff exactly)
$ diff a.txt a.txt
$

# Default: ed-style, positional. Line N of FILE1 vs line N of FILE2; each
# differing position emits `< ` (FILE1) then `> ` (FILE2).
$ diff a.txt b.txt          # a.txt: apple/banana/cherry  b.txt: banana/cherry/kiwi
< apple
> banana
< banana
> cherry
< cherry
> kiwi

# A single changed middle line: only that position differs.
$ diff a.txt b.txt          # a.txt: a/b/c  b.txt: a/x/c
< b
> x

# FILE1 longer: trailing line has no counterpart, emits only `< `.
$ diff a.txt b.txt          # a.txt: a/b  b.txt: a
< b

# Unified (-u): `--- a` / `+++ b` headers, one `@@ -1,N +1,M @@` hunk, then
# ` `/`-`/`+` body rows for the whole span.
$ diff -u a.txt b.txt       # a.txt: a/b/c  b.txt: a/x/c
--- a
+++ b
@@ -1,3 +1,3 @@
 a
-b
+x
 c

# --unified is the long form of -u (identical output).
$ diff --unified a.txt b.txt
--- a
+++ b
@@ -1,3 +1,3 @@
 a
-b
+x
 c
```

## Intentional Divergences

- Comparison is **positional**: line N of FILE1 is compared to line N of FILE2. GNU `diff` computes a longest-common-subsequence (LCS) edit script, so on inserts/deletes in the middle of a file the hunks and counts differ from GNU. This is the documented behavior of this module, not a bug.
- The default output is ed-style `< line` / `> line` rows. GNU `diff`'s default is a normal diff with `1,3c1,3`-style change commands; there is no GNU mode that produces these exact bytes.
- The unified diff (`-u`) emits exactly one hunk spanning both files (`@@ -1,N +1,M @@`) rather than the minimal context-bounded hunks GNU produces, and its `---`/`+++` headers carry the labels `a` / `b` rather than filenames and timestamps.
- Only identical files match GNU exactly: no output and exit status 0. When files differ, the exit status is 1 (the standard "differences found" signal, not an error).
