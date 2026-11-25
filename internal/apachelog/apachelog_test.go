package apachelog

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"testing/iotest"
)

func open(t *testing.T, name string) *os.File {
	t.Helper()
	f, err := os.Open("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

func TestParseLine(t *testing.T) {
	tests := []struct {
		name, line string
		ok         bool
		level, msg string
	}{
		{"valid", "[Mon Jan 02 10:00:00 2023] [error] hello world", true, "error", "hello world"},
		{"empty message", "[Mon Jan 02 10:00:00 2023] [error] ", true, "error", ""},
		{"brackets in message", "[Mon Jan 02 10:00:00 2023] [notice] a [b] c", true, "notice", "a [b] c"},
		{"empty", "", false, "", ""},
		{"no prefix", "Mon Jan 02 10:00:00 2023] [error] x", false, "", ""},
		{"bad time", "[Foo Bar 99 10:00:00 2023] [error] x", false, "", ""},
		{"empty level", "[Mon Jan 02 10:00:00 2023] [] x", false, "", ""},
		{"digit in level", "[Mon Jan 02 10:00:00 2023] [e1] x", false, "", ""},
		{"no level bracket", "[Mon Jan 02 10:00:00 2023] error x", false, "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, ok := ParseLine(tc.line)
			if ok != tc.ok {
				t.Fatalf("ok=%v want %v", ok, tc.ok)
			}
			if ok && (r.Level != tc.level || r.Message != tc.msg) {
				t.Fatalf("got level=%q msg=%q", r.Level, r.Message)
			}
		})
	}
}

func TestParseFixtures(t *testing.T) {
	tests := []struct {
		file                    string
		total, parsed, unparsed int
		unparsedLines           []int
	}{
		{"valid.log", 3, 3, 0, nil},
		{"malformed.log", 6, 2, 4, []int{2, 3, 5, 6}},
		{"empty.log", 0, 0, 0, nil},
		{"no_newline.log", 1, 1, 0, nil},
		{"crlf.log", 2, 2, 0, nil},
	}
	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			res, err := Parse(open(t, tc.file), 0)
			if err != nil {
				t.Fatal(err)
			}
			if !res.Complete {
				t.Error("expected Complete")
			}
			if res.TotalLines != tc.total || res.Parsed != tc.parsed || res.Unparsed != tc.unparsed {
				t.Fatalf("got total=%d parsed=%d unparsed=%d", res.TotalLines, res.Parsed, res.Unparsed)
			}
			if fmt.Sprint(res.UnparsedLines) != fmt.Sprint(tc.unparsedLines) && len(tc.unparsedLines) > 0 {
				t.Fatalf("unparsed lines %v want %v", res.UnparsedLines, tc.unparsedLines)
			}
			if res.TotalLines != res.Parsed+res.Unparsed {
				t.Fatal("invariant total = parsed + unparsed violated")
			}
		})
	}
}

func TestCRLFStripped(t *testing.T) {
	res, err := Parse(open(t, "crlf.log"), 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := res.Records[1].Message; got != "crlf two" {
		t.Fatalf("message %q", got)
	}
}

func TestOversizedLineIsFatal(t *testing.T) {
	res, err := Parse(open(t, "oversized.log"), 1024)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrLineTooLong) {
		t.Fatalf("want ErrLineTooLong, got %v", err)
	}
	var re *ReadError
	if !errors.As(err, &re) || re.Line != 2 {
		t.Fatalf("want ReadError at line 2, got %v", err)
	}
	if res.Complete {
		t.Error("must not be Complete")
	}
	if res.Parsed != 1 {
		t.Errorf("parsed before failure = %d", res.Parsed)
	}
}

func TestOversizedAcceptedWithLargerLimit(t *testing.T) {
	res, err := Parse(open(t, "oversized.log"), 8192)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Records[1].Message) != 3000 {
		t.Fatalf("message truncated: %d", len(res.Records[1].Message))
	}
}

func TestLineExactlyAtLimit(t *testing.T) {
	line := "[Mon Jan 02 10:00:00 2023] [error] " + strings.Repeat("x", 100)
	for _, limit := range []int{len(line), len(line) - 1} {
		_, err := Parse(strings.NewReader(line+"\r\n"), limit)
		if (limit == len(line)) != (err == nil) {
			t.Errorf("limit %d: err=%v", limit, err)
		}
	}
}

type failingReader struct{ data string }

func (f *failingReader) Read(p []byte) (int, error) {
	if f.data == "" {
		return 0, errors.New("disk on fire")
	}
	n := copy(p, f.data)
	f.data = f.data[n:]
	return n, nil
}

func TestReaderErrorIsFatalNotMalformed(t *testing.T) {
	res, err := Parse(&failingReader{data: "[Mon Jan 02 10:00:00 2023] [error] a\n[Mon Jan 02 10:00:01 2023] [error] b\n"}, 0)
	if err == nil || !strings.Contains(err.Error(), "disk on fire") {
		t.Fatalf("got %v", err)
	}
	if res.Complete || res.Parsed != 2 || res.Unparsed != 0 {
		t.Fatalf("unexpected result %+v", res)
	}
	if _, err := Parse(iotest.ErrReader(errors.New("boom")), 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestUnparsedListCapped(t *testing.T) {
	res, err := Parse(strings.NewReader(strings.Repeat("junk\n", 150)), 0)
	if err != nil {
		t.Fatal(err)
	}
	if res.Unparsed != 150 || len(res.UnparsedLines) != 100 || !res.UnparsedTruncated {
		t.Fatalf("unparsed=%d listed=%d trunc=%v", res.Unparsed, len(res.UnparsedLines), res.UnparsedTruncated)
	}
}

func TestMaskNumbers(t *testing.T) {
	tests := map[string]string{
		"worker 12 failed on port 8080": "worker <N> failed on port <N>",
		"jk2_init() Found child 6725":   "jk<N>_init() Found child <N>",
		"no digits":                     "no digits",
		"":                              "",
	}
	for in, want := range tests {
		if got := MaskNumbers(in); got != want {
			t.Errorf("MaskNumbers(%q)=%q want %q", in, got, want)
		}
	}
}

func TestSummarize(t *testing.T) {
	res, err := Parse(open(t, "valid.log"), 0)
	if err != nil {
		t.Fatal(err)
	}
	s := Summarize(res, 10)
	if s.FirstTimestamp != "2023-01-02T09:59:00" || s.LastTimestamp != "2023-01-02T10:00:05" {
		t.Errorf("timestamps %s %s", s.FirstTimestamp, s.LastTimestamp)
	}
	if len(s.LevelCounts) != 2 || s.LevelCounts[0] != (Count{"error", 2}) {
		t.Errorf("levels %v", s.LevelCounts)
	}
	if s.TopPatterns[0] != (Count{"worker <N> failed on port <N>", 2}) {
		t.Errorf("patterns %v", s.TopPatterns)
	}
	if got := Summarize(res, 1); len(got.TopPatterns) != 1 {
		t.Errorf("top limit not applied")
	}
	if s.TotalLines != s.Parsed+s.Unparsed {
		t.Error("invariant")
	}
}

func TestSummarizeEmpty(t *testing.T) {
	res, _ := Parse(open(t, "empty.log"), 0)
	s := Summarize(res, 10)
	if s.FirstTimestamp != "" || s.LastTimestamp != "" || len(s.LevelCounts) != 0 {
		t.Fatalf("%+v", s)
	}
}

func TestDeterministicOrdering(t *testing.T) {
	// Ties on count must be broken by key ascending, regardless of map order.
	var b strings.Builder
	for _, w := range []string{"zeta", "alpha", "mike", "bravo", "yankee", "charlie"} {
		fmt.Fprintf(&b, "[Mon Jan 02 10:00:00 2023] [%s] %s event\n", w, w)
	}
	res, _ := Parse(strings.NewReader(b.String()), 0)
	first := fmt.Sprint(Summarize(res, 10))
	for i := 0; i < 50; i++ {
		if got := fmt.Sprint(Summarize(res, 10)); got != first {
			t.Fatal("summary not deterministic")
		}
	}
	s := Summarize(res, 10)
	for i := 1; i < len(s.LevelCounts); i++ {
		if s.LevelCounts[i-1].Key >= s.LevelCounts[i].Key {
			t.Fatalf("ties not sorted by key: %v", s.LevelCounts)
		}
	}
}

func TestReferenceAgreement(t *testing.T) {
	res, _ := Parse(open(t, "valid.log"), 0)
	rows, err := ReadReference(open(t, "valid_ref.csv"))
	if err != nil {
		t.Fatal(err)
	}
	a := Compare(res, rows)
	if !a.Aligned || a.Compared != 3 || a.TimeMatches != 3 || a.ContentMatches != 3 || a.LevelMatches != 2 || a.AllFieldMatches != 2 {
		t.Fatalf("%+v", a)
	}
}

func TestReferenceAlignmentFailures(t *testing.T) {
	res, _ := Parse(open(t, "valid.log"), 0)
	good := []RefRow{{1, "", "", ""}, {2, "", "", ""}, {3, "", "", ""}}
	tests := map[string][]RefRow{
		"short":      good[:2],
		"bad lineid": {{1, "", "", ""}, {3, "", "", ""}, {2, "", "", ""}},
	}
	for name, rows := range tests {
		if a := Compare(res, rows); a.Aligned || a.Compared != 0 || a.AlignmentReason == "" {
			t.Errorf("%s: %+v", name, a)
		}
	}
	mal, _ := Parse(open(t, "malformed.log"), 0)
	if a := Compare(mal, make([]RefRow, mal.TotalLines)); a.Aligned {
		t.Error("unparsed lines must break alignment")
	}
	if a := Compare(&Result{}, nil); a.Aligned {
		t.Error("incomplete read must not align")
	}
}

func TestReadReferenceErrors(t *testing.T) {
	for name, in := range map[string]string{
		"bad header":  "a,b,c,d,e,f\n",
		"bad lineid":  "LineId,Time,Level,Content,EventId,EventTemplate\nx,t,l,c,e,t\n",
		"short row":   "LineId,Time,Level,Content,EventId,EventTemplate\n1,t,l\n",
		"empty":       "",
		"bad quoting": "LineId,Time,Level,Content,EventId,EventTemplate\n1,t,l,\"c,e,t\n",
	} {
		if _, err := ReadReference(strings.NewReader(in)); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestReferenceQuotedFields(t *testing.T) {
	in := "LineId,Time,Level,Content,EventId,EventTemplate\n1,t,l,\"a, \"\"b\"\"\",E1,x\n"
	rows, err := ReadReference(strings.NewReader(in))
	if err != nil || rows[0].Content != `a, "b"` {
		t.Fatalf("%v %v", rows, err)
	}
}
