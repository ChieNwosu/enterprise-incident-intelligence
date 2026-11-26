package apachelog

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// RefRow is one row of the Loghub structured CSV. EventId and EventTemplate
// are log-pattern reference fields, not incident or anomaly labels, and are
// deliberately not used here.
type RefRow struct {
	LineID  int
	Time    string
	Level   string
	Content string
}

// ReadReference reads a Loghub-style structured CSV with header
// LineId,Time,Level,Content,EventId,EventTemplate.
func ReadReference(r io.Reader) ([]RefRow, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = 6
	header, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("reference csv header: %w", err)
	}
	want := []string{"LineId", "Time", "Level", "Content", "EventId", "EventTemplate"}
	for i, w := range want {
		if header[i] != w {
			return nil, fmt.Errorf("reference csv header column %d is %q, want %q", i, header[i], w)
		}
	}
	var rows []RefRow
	for {
		rec, err := cr.Read()
		if errors.Is(err, io.EOF) {
			return rows, nil
		}
		if err != nil {
			return nil, fmt.Errorf("reference csv: %w", err)
		}
		id, err := strconv.Atoi(rec[0])
		if err != nil {
			return nil, fmt.Errorf("reference csv row %d: bad LineId %q", len(rows)+1, rec[0])
		}
		rows = append(rows, RefRow{LineID: id, Time: rec[1], Level: rec[2], Content: rec[3]})
	}
}

// Agreement reports parsed-field agreement against the reference CSV.
type Agreement struct {
	Aligned         bool   `json:"aligned"`
	AlignmentReason string `json:"alignment_reason"`
	ReferenceRows   int    `json:"reference_rows"`
	Compared        int    `json:"compared_rows"`
	TimeMatches     int    `json:"time_matches"`
	LevelMatches    int    `json:"level_matches"`
	ContentMatches  int    `json:"content_matches"`
	AllFieldMatches int    `json:"all_field_matches"`
	Note            string `json:"note"`
}

// Compare checks row alignment and, only if it holds, compares fields.
// Alignment requires: input was fully read, every line parsed (so parsed
// record i is physical line i), reference LineId values are exactly 1..n,
// and the reference row count equals the input line count.
func Compare(res *Result, ref []RefRow) Agreement {
	a := Agreement{
		ReferenceRows: len(ref),
		Note:          "Compares Time, Level, Content only. EventId and EventTemplate are pattern reference fields, not incident or anomaly labels, and are not evaluated.",
	}
	switch {
	case !res.Complete:
		a.AlignmentReason = "input read was not complete"
		return a
	case res.Unparsed != 0:
		a.AlignmentReason = "unparsed lines present, so record index cannot be assumed to equal line number"
		return a
	case len(ref) != res.TotalLines:
		a.AlignmentReason = fmt.Sprintf("row count mismatch: reference %d, input lines %d", len(ref), res.TotalLines)
		return a
	}
	for i, row := range ref {
		if row.LineID != i+1 {
			a.AlignmentReason = fmt.Sprintf("reference LineId not sequential at row %d (got %d)", i+1, row.LineID)
			return a
		}
	}
	a.Aligned = true
	a.AlignmentReason = "line counts equal, all lines parsed, reference LineId sequential 1..n"
	for i, row := range ref {
		rec := res.Records[i]
		t := rec.RawTime == row.Time
		l := rec.Level == row.Level
		c := rec.Message == row.Content
		a.Compared++
		if t {
			a.TimeMatches++
		}
		if l {
			a.LevelMatches++
		}
		if c {
			a.ContentMatches++
		}
		if t && l && c {
			a.AllFieldMatches++
		}
	}
	return a
}
