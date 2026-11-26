# bufio.Scanner token limit

**Concept**: `bufio.Scanner` stops with `bufio.ErrTooLong` when a line exceeds its buffer maximum (default 64 KiB). `Scan()` then returns false, so a loop that ignores `Err()` looks like a normal end of input and silently drops the rest of the file.

**Practice**: set `Buffer(buf, max)` explicitly, always check `Scanner.Err()` after the loop, and treat too-long records as a fatal error rather than truncating.

**Provenance**: Go standard library documentation for `bufio.Scanner` and `bufio.ErrTooLong`.

**Evidence**: `internal/apachelog/parse.go` (`Parse`), tests `TestOversizedLineIsFatal`, `TestLineExactlyAtLimit`, `TestReaderErrorIsFatalNotMalformed`; [CASE-001](../../support/cases/CASE-001/README.md).
