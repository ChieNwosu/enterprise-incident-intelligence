# Numeric masking is a heuristic

**Concept**: replacing digit runs with a placeholder groups messages that differ only in numbers (PIDs, ports, slots). It is cheap and deterministic but not a validated log-template classification algorithm. It over-merges (different meanings, same digits-free text) and under-merges (variable text that is not numeric).

**Practice**: label such output as heuristic, keep ordering deterministic (count descending, key ascending), and do not treat it as ground truth.

**Provenance**: general log-parsing practice; the Loghub `EventTemplate` column comes from a separate benchmark process and is a pattern reference, not an incident or anomaly label.

**Evidence**: `internal/apachelog/summary.go` (`MaskNumbers`, `Summarize`), tests `TestMaskNumbers`, `TestDeterministicOrdering`. On Apache 2k the masking yields 6 patterns ([day1-run.md](../../docs/evidence/day1-run.md)).
