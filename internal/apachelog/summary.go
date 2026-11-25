package apachelog

import (
	"regexp"
	"sort"
	"time"
)

// Count is a key with its occurrence count.
type Count struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

// Summary is the JSON-serialisable report. Field order is fixed by the struct.
type Summary struct {
	TotalLines        int     `json:"total_lines"`
	Parsed            int     `json:"parsed"`
	Unparsed          int     `json:"unparsed"`
	UnparsedLines     []int   `json:"unparsed_line_numbers"`
	UnparsedTruncated bool    `json:"unparsed_line_numbers_truncated"`
	FirstTimestamp    string  `json:"first_timestamp"`
	LastTimestamp     string  `json:"last_timestamp"`
	TimestampNote     string  `json:"timestamp_note"`
	LevelCounts       []Count `json:"level_counts"`
	TopPatterns       []Count `json:"top_patterns"`
	PatternNote       string  `json:"pattern_note"`
}

const (
	timestampNote = "Wall-clock times copied from the log; the original timezone is unknown and no zone is assumed."
	patternNote   = "Patterns come from a numeric-masking heuristic (digit runs replaced by <N>); not a validated template classification."
)

var digits = regexp.MustCompile(`[0-9]+`)

// MaskNumbers replaces each run of ASCII digits with <N>. Heuristic only.
func MaskNumbers(s string) string { return digits.ReplaceAllString(s, "<N>") }

// sortCounts orders by count descending, then key ascending.
func sortCounts(c []Count) {
	sort.Slice(c, func(i, j int) bool {
		if c[i].Count != c[j].Count {
			return c[i].Count > c[j].Count
		}
		return c[i].Key < c[j].Key
	})
}

func toCounts(m map[string]int) []Count {
	out := make([]Count, 0, len(m))
	for k, v := range m {
		out = append(out, Count{k, v})
	}
	sortCounts(out)
	return out
}

// Summarize builds a Summary keeping at most topN patterns (topN < 0 means 10).
func Summarize(res *Result, topN int) Summary {
	if topN < 0 {
		topN = 10
	}
	s := Summary{
		TotalLines:        res.TotalLines,
		Parsed:            res.Parsed,
		Unparsed:          res.Unparsed,
		UnparsedLines:     append([]int{}, res.UnparsedLines...),
		UnparsedTruncated: res.UnparsedTruncated,
		TimestampNote:     timestampNote,
		PatternNote:       patternNote,
	}
	levels := map[string]int{}
	patterns := map[string]int{}
	var first, last time.Time
	for i, r := range res.Records {
		levels[r.Level]++
		patterns[MaskNumbers(r.Message)]++
		if i == 0 || r.Time.Before(first) {
			first = r.Time
		}
		if i == 0 || r.Time.After(last) {
			last = r.Time
		}
	}
	if len(res.Records) > 0 {
		s.FirstTimestamp = first.Format(NaiveLayout)
		s.LastTimestamp = last.Format(NaiveLayout)
	}
	s.LevelCounts = toCounts(levels)
	s.TopPatterns = toCounts(patterns)
	if len(s.TopPatterns) > topN {
		s.TopPatterns = s.TopPatterns[:topN]
	}
	return s
}
