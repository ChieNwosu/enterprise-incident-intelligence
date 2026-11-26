# enterprise-incident-intelligence

Cloud-native incident intelligence platform integrating Go, SAS, PyTorch, and AI-assisted enterprise troubleshooting, developed through a self-paced engineering curriculum (Enterprise Incident Intelligence Academy).

## Status

Only Day 1 is implemented. Everything else is planned.

| Area | Status |
| --- | --- |
| Go module, `eii` CLI, Apache error-log parser and summary (JSON) | Done (Day 1) |
| Reference CSV field agreement check | Done (Day 1) |
| Synthetic test fixtures, unit and CLI tests | Done (Day 1) |
| Dataset download script and provenance notes | Done (Day 1) |
| GitHub Actions workflow, devcontainer | Written. CI for commit 452f848 (pre-remediation): workflow `ci` run [45618163063](https://github.com/ChieNwosu/enterprise-incident-intelligence/actions/runs/45618163063) (push) and run [45618184731](https://github.com/ChieNwosu/enterprise-incident-intelligence/actions/runs/45618184731) (pull_request) both concluded success, per the GitHub REST API. Per-step job logs were not inspected. This applies only to 452f848. CI for approved engineering commit f6e83b3: workflow `ci` run [45747243041](https://github.com/ChieNwosu/enterprise-incident-intelligence/actions/runs/45747243041) (push) and run [45747250736](https://github.com/ChieNwosu/enterprise-incident-intelligence/actions/runs/45747250736) (pull_request) both completed with success, and the gofmt, vet, test and build steps each concluded success, per the GitHub REST API (per-step conclusions only, raw job logs not read). CI for the later documentation-only commit is to be checked after the push. |
| Curriculum skeleton, evergreen notes, support cases CASE-001 and CASE-002 | Done (Day 1), skeleton level |
| AWS, Kubernetes, PyTorch, SAS, anomaly or incident models | Planned, not started |

## Review status and project management

Independent Codex review of the approved engineering commit f6e83b36b1791277e787cf8dfb88ef74df2fc0b1 (REM-001, the CASE-002 fix) returned APPROVED with no remaining merge blockers. PR #1 is open and not merged. GitHub Issues, GitHub Projects and Markdown are the canonical project management tools. Jira, Confluence and MCP-based project management automation are deferred.

## Quick start

Requires Go 1.25.

```bash
go vet ./... && go test -race ./... && go build ./...
bash scripts/download_loghub_apache.sh        # fetches into data/raw/ (gitignored)
go run ./cmd/eii -input data/raw/Apache_2k.log \
  -reference data/raw/Apache_2k.log_structured.csv -output data/out.json
```

Flags: `-input` (required), `-reference`, `-output`, `-top`, `-max-line-bytes`, `-strict`.
`-output` must be a different file from both `-input` and `-reference`; otherwise the run exits 1 and the sources are left untouched. A report from an earlier successful run can remain on disk after a later fatal run. With `-strict`, a valid report is written before the exit code 3 is returned.
Exit codes: 0 success (malformed records are counted, not fatal), 1 fatal input/output error (including a record longer than the limit), 2 usage error, 3 malformed records present with `-strict`.

## Data

The Loghub Apache 2k sample is third-party data and is never committed. See [docs/dataset-provenance.md](docs/dataset-provenance.md). `EventId` and `EventTemplate` in the reference CSV are log-pattern reference fields, not incident or anomaly labels.

## Documents

- [docs/setup-browser.md](docs/setup-browser.md): browser-only setup with Codespaces
- [docs/dataset-provenance.md](docs/dataset-provenance.md)
- [docs/parser-limitations.md](docs/parser-limitations.md)
- [docs/evidence/day1-run.md](docs/evidence/day1-run.md): counts from the real run
- [curriculum/README.md](curriculum/README.md)
- [knowledge/evergreen/](knowledge/evergreen/)
- [support/cases/CASE-001/README.md](support/cases/CASE-001/README.md)

## License

See [LICENSE](LICENSE). Third-party datasets carry their own terms (see provenance).
