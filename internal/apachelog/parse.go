// Package apachelog parses Apache httpd error-log lines of the form
//
//	[Sun Dec 04 04:47:44 2005] [notice] message text
//
// and builds a deterministic summary. Standard library only.
package apachelog

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// DefaultMaxLineBytes is the longest single record accepted by default.
// Longer records cause a fatal error; they are never truncated.
const DefaultMaxLineBytes = 1 << 20

const tsLayout = "Mon Jan 02 15:04:05 2006"

// NaiveLayout is the output layout for timestamps. It carries no zone
// designator because the zone of the original log is unknown.
const NaiveLayout = "2006-01-02T15:04:05"

// Record is one successfully parsed line.
type Record struct {
	LineNo int
	// RawTime is the timestamp text exactly as it appeared between brackets.
	RawTime string
	// Time is the wall-clock time. Its location (UTC) is a carrier only and
	// does not mean the log was written in UTC.
	Time    time.Time
	Level   string
	Message string
}

// ParseLine parses a single line (without line terminator).
func ParseLine(line string) (Record, bool) {
	var r Record
	if !strings.HasPrefix(line, "[") {
		return r, false
	}
	end := strings.Index(line, "] [")
	if end < 0 {
		return r, false
	}
	rawTime := line[1:end]
	t, err := time.Parse(tsLayout, rawTime)
	if err != nil {
		return r, false
	}
	rest := line[end+3:]
	lvEnd := strings.Index(rest, "] ")
	if lvEnd <= 0 {
		return r, false
	}
	level := rest[:lvEnd]
	for _, c := range level {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
			return r, false
		}
	}
	return Record{RawTime: rawTime, Time: t, Level: level, Message: rest[lvEnd+2:]}, true
}

// Result holds the outcome of reading an input.
type Result struct {
	Records []Record
	// TotalLines counts every line read, parsed or not.
	TotalLines int
	Parsed     int
	Unparsed   int
	// UnparsedLines lists line numbers of malformed records (capped).
	UnparsedLines []int
	// UnparsedTruncated is true when UnparsedLines hit its cap.
	UnparsedTruncated bool
	// Complete is true only if the whole input was read without fatal error.
	Complete bool
}

const maxUnparsedListed = 100

// ErrLineTooLong is returned (wrapped) when a record exceeds the size limit.
var ErrLineTooLong = errors.New("input record exceeds maximum line size")

// ReadError wraps a fatal I/O error with the line position reached.
type ReadError struct {
	Line int // number of the line being read when the error occurred
	Err  error
}

func (e *ReadError) Error() string { return fmt.Sprintf("read failed at line %d: %v", e.Line, e.Err) }
func (e *ReadError) Unwrap() error { return e.Err }

// Parse reads all lines from r. Malformed lines are counted as unparsed and
// are not an error. A fatal read error (including a line longer than
// maxLineBytes) returns a non-nil error together with the partial Result,
// whose Complete field is false. maxLineBytes <= 0 selects the default.
func Parse(r io.Reader, maxLineBytes int) (*Result, error) {
	if maxLineBytes <= 0 {
		maxLineBytes = DefaultMaxLineBytes
	}
	res := &Result{}
	sc := bufio.NewScanner(r)
	// Scanner needs room for the line plus a possible trailing "\r" and "\n".
	sc.Buffer(make([]byte, 0, 64*1024), maxLineBytes+2)
	for sc.Scan() {
		res.TotalLines++
		line := strings.TrimSuffix(sc.Text(), "\r")
		if len(line) > maxLineBytes {
			return res, &ReadError{Line: res.TotalLines, Err: ErrLineTooLong}
		}
		rec, ok := ParseLine(line)
		if !ok {
			res.Unparsed++
			if len(res.UnparsedLines) < maxUnparsedListed {
				res.UnparsedLines = append(res.UnparsedLines, res.TotalLines)
			} else {
				res.UnparsedTruncated = true
			}
			continue
		}
		rec.LineNo = res.TotalLines
		res.Parsed++
		res.Records = append(res.Records, rec)
	}
	if err := sc.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			err = ErrLineTooLong
		}
		return res, &ReadError{Line: res.TotalLines + 1, Err: err}
	}
	if res.TotalLines != res.Parsed+res.Unparsed {
		return res, fmt.Errorf("internal invariant violated: total %d != parsed %d + unparsed %d",
			res.TotalLines, res.Parsed, res.Unparsed)
	}
	res.Complete = true
	return res, nil
}
