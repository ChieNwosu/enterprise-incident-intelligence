# CASE-002: `-output` can overwrite the input log or reference CSV

Reported by: independent Codex review of PR #1 (reviewed commit 452f848). Severity: high (data loss).

## Defect

`eii -output PATH` accepted a path that identified the input log or the reference CSV. The source file was replaced with JSON and the run exited 0.

## Disposable reproduction (before the fix)

```bash
d=$(mktemp -d)
cp cmd/eii/testdata/valid.log $d/in.log
go build -o $d/eii ./cmd/eii
$d/eii -input $d/in.log -output $d/in.log; echo "exit=$?"
head -c 20 $d/in.log
```

Before the fix this exited 0 and `in.log` began with JSON. The same applied to `-reference`, relative aliases, symlinks and hard links. In the regression tests, all seven alias cases failed against the old `main.go`.

## Expected behavior

Exit 1 with a fatal diagnostic, and both source files unchanged. Distinct destinations, including ones that do not exist yet, still work.

## Root cause

`run` passed `-output` straight to `os.WriteFile`, which truncates the target. It never compared the destination with the files it had opened.

## Correction

`cmd/eii/main.go`: the input and reference files are stat-ed after opening. Before writing a named output, `checkOutput` calls `os.Stat` on the destination (this follows symlinks) and compares it with each source using `os.SameFile`. A not-exist error is accepted; any other stat error is fatal. Stdout behavior and strict-mode semantics are unchanged.

## Regression tests (`cmd/eii/main_test.go`, synthetic fixtures in `t.TempDir()`)

- `TestOutputAliasesRejected`: output equals input, equals reference, relative alias, symlink to input and to reference, hard link to input and to reference. Each asserts exit 1, the diagnostic, empty stdout and byte-identical sources. Symlink and hard-link cases skip if the platform refuses to create links.
- `TestOutputDistinctAccepted`: new and existing distinct destinations succeed.
- `TestStrictWritesReportThenExits3`: malformed input with `-strict` and a valid output writes valid JSON and exits 3.
- `TestOutputIsDirectory`: a directory as output exits 1 and the source is unchanged.

## Actual results

All of the above passed on Linux with Go 1.25.1 (`go test -race -count=1 ./...`, plus gofmt, vet and build clean). The Apache 2k sample still gives input 2000, parsed 2000, unparsed 0, notice 1405, error 595, and Time, Level and Content each matching 2000 of 2000. Pointing `-output` at the downloaded sample log exited 1 and the dataset checksums still verified. CI for the fix commit is pending until checked.

## Remaining limitations

- The check and write are not atomic (time-of-check to time-of-use). Concurrent filesystem changes by another process can defeat it.
- Write is not atomic (no temp-file and rename); a failed write can leave a partial output.
- Only files opened by this run are compared. Other sensitive files are not protected.
- Hard-link and symlink behavior was tested on Linux only.
