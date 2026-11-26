# Day 1 run evidence (counts only)

No raw log content is recorded here. Independent Codex review approved engineering commit f6e83b36b1791277e787cf8dfb88ef74df2fc0b1 (REM-001); this document's own commit is documentation-only. The local results below are not CI results. CI for commit 452f848 (pre-remediation): workflow `ci` run [45618163063](https://github.com/ChieNwosu/enterprise-incident-intelligence/actions/runs/45618163063) (push) and run [45618184731](https://github.com/ChieNwosu/enterprise-incident-intelligence/actions/runs/45618184731) (pull_request) both concluded success, per the GitHub REST API. Per-step job logs were not inspected. This applies only to 452f848. CI for approved engineering commit f6e83b3: workflow `ci` run [45747243041](https://github.com/ChieNwosu/enterprise-incident-intelligence/actions/runs/45747243041) (push) and run [45747250736](https://github.com/ChieNwosu/enterprise-incident-intelligence/actions/runs/45747250736) (pull_request) both completed with success, and the gofmt, vet, test and build steps each concluded success, per the GitHub REST API (per-step conclusions only, raw job logs not read). CI for the later documentation-only commit is to be checked after the push.

## Local verification (Go 1.25.1 linux/amd64)

| Command | Result |
| --- | --- |
| `go vet ./...` | exit 0 |
| `go test -race ./...` | ok for `cmd/eii` and `internal/apachelog` |
| `go build ./...` | exit 0 |

## Parser on Loghub Apache 2k (pinned commit dd61d09...)

Command: `go run ./cmd/eii -input data/raw/Apache_2k.log -reference data/raw/Apache_2k.log_structured.csv`

| Metric | Value |
| --- | --- |
| input (total_lines) | 2000 |
| parsed | 2000 |
| unparsed | 0 |
| total = parsed + unparsed | yes |
| level counts | notice 1405, error 595 |
| distinct masked patterns | 6 |
| first / last timestamp (zone unknown) | 2005-12-04T04:47:44 / 2005-12-05T19:15:57 |
| reference alignment | aligned (2000 rows, LineId 1..2000) |
| Time / Level / Content matches | 2000 / 2000 / 2000 of 2000 |
| all three fields match | 2000 |

SHA-256 of inputs: see [dataset-provenance.md](../dataset-provenance.md).

Agreement here means the parsed Time, Level and Content equal the reference CSV columns. It says nothing about EventId or EventTemplate.

## Remediation verification (commit f6e83b3)

- Eight source-file alias combinations were rejected in independent review. The automated suite has seven alias regression subtests (output equal to input, output equal to reference, a relative alias, a symlink to each source, a hard link to each source), passing without skips. All seven failed against the original implementation.
- CASE-001 and CASE-002 were independently reproduced. Both source files were preserved byte for byte during disposable collision tests.
- gofmt, vet, race-enabled tests and build passed.

## Sprint 0 retrospective

### What went well

- 2,000 Apache records parsed with zero unparsed (1,405 notice, 595 error), and 2,000 exact text matches for each of Time, Level and Content against the reference. Six masked patterns.
- Defects were recorded as support cases (CASE-001, CASE-002) with reproductions.

### Why the original tests missed the overwrite

The original tests covered parsing, summary, reference agreement and exit codes. None exercised the `-output` path against the source files, so the overwrite of the input or reference file went untested. Independent Codex review found it.

### How review and regression testing helped

Independent review reproduced the defect. The fix added seven regression subtests, all of which fail on the original implementation and pass on the fix, so they demonstrably detect the defect.

### Test-isolation deviation and prevention

During REM-001 the coder agent ran one direct CLI command with `-output` pointing at `data/raw/Apache_2k.log`, the real downloaded dataset, to confirm the guard. It was rejected with exit 1 and checksums were verified afterwards, so nothing changed. This went against the instruction to keep destructive tests off the downloaded dataset. Prevention: verify guards only on copies or disposable fixtures under `t.TempDir()` or a temp directory, and consider making `data/raw` read-only during manual checks.

### Remaining technical debt (not resolved in Sprint 0)

- Time agreement compares raw timestamp text.
- An empty log with a header-only CSV reports `aligned: true` with 0 rows compared.
- The summary JSON lists the top N patterns with no distinct-pattern total.
- CI actions are pinned to major tags, not SHAs.
- Untested edge cases: a CRLF record at the size limit plus one, a CSV with a byte-order mark, and an unwritable `-output` path (the directory-as-output test covers only that one case).
- Output writes are not atomic, and the check-then-write sequence has a race (TOCTOU).
- Records are held in memory (no streaming).
- JSON output contains masked message content, with no comprehensive redaction.
- Symlink and hard-link tests run only on Linux.

### Student verification lab

Not completed. The student has not yet done the verification lab.

### Project management decision

GitHub Issues, GitHub Projects and Markdown are canonical. Jira, Confluence and MCP-based project management automation are deferred.
