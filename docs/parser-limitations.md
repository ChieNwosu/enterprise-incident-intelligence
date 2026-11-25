# Parser limitations and behaviour

- **Format**: only `[Day Mon DD HH:MM:SS YYYY] [level] message` is recognised. Anything else (blank lines, other formats, bad timestamps, levels with non-letters) is counted as unparsed. Line numbers of the first 100 malformed lines are reported; their content never is.
- **Malformed vs fatal**: malformed records are normal data and do not fail the run. Fatal errors are open/read/write failures and records over the size limit. Fatal errors exit 1 and produce no summary.
- **Oversized records**: the limit is `-max-line-bytes` (default 1 MiB). A longer record stops the run with `ErrLineTooLong`. It is never truncated and never counted as merely malformed. `bufio.Scanner.Err()` is checked explicitly.
- **Invariant**: for a completed read, `total_lines = parsed + unparsed`. A violation returns an internal error.
- **Line endings**: LF and CRLF are accepted; a final line without a terminator is read. A lone CR in the middle of a line is kept as content.
- **Timestamps**: the log has no timezone. Times are emitted as `YYYY-MM-DDTHH:MM:SS` with no zone suffix, and `timestamp_note` says so. Internally they are held in a UTC location only as a carrier, so do not compare them to zoned times. The weekday text is not checked against the date.
- **Patterns**: digit runs are replaced with `<N>`. This is a heuristic. It merges messages that differ only in digits and splits messages that differ in other variable text (paths, hostnames). It is not a validated template-classification algorithm and is not compared to `EventTemplate`.
- **Ordering**: counts sort by count descending, then key ascending, so output is deterministic.
- **Reference agreement**: fields are compared only when alignment is validated (complete read, zero unparsed lines, equal row and line counts, `LineId` is 1..n). Otherwise `aligned` is false with a reason and no field counts are given.
- **Memory**: all parsed records are held in memory. Fine for 2k lines, not designed for multi-GB logs.
- **Privacy**: the JSON contains masked message patterns, which are derived from log content. Treat output files as sensitive for real logs. CI runs only synthetic fixtures and uploads no artifacts.
