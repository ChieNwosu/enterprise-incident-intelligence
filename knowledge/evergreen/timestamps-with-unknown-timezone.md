# Timestamps with unknown timezone

**Concept**: a log timestamp without a zone is a wall-clock reading. Labelling it UTC or local invents information and can shift events by hours when merged with zoned sources.

**Practice**: parse to a naive value, emit it without a zone suffix, and state the limitation in the output. Do not compare it with zoned times until the zone is established.

**Provenance**: Go `time` package behaviour (`time.Parse` without a zone yields UTC); Apache error log lines carry no zone.

**Evidence**: `internal/apachelog/parse.go` (`NaiveLayout`), `Summary.TimestampNote`, [parser-limitations.md](../../docs/parser-limitations.md).
