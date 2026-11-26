# Day 1 run evidence (counts only)

No raw log content is recorded here. The local results below are not CI results. CI for commit 452f848 (pre-remediation): workflow `ci` run [45618163063](https://github.com/ChieNwosu/enterprise-incident-intelligence/actions/runs/45618163063) (push) and run [45618184731](https://github.com/ChieNwosu/enterprise-incident-intelligence/actions/runs/45618184731) (pull_request) both concluded success, per the GitHub REST API. Per-step job logs were not inspected. This applies only to 452f848; CI for later commits is pending until checked.

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
