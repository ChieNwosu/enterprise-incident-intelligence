// Command eii parses an Apache error log and writes a JSON summary.
//
// Exit codes: 0 success (malformed records are reported, not fatal),
// 1 fatal input/output error, 2 usage error, 3 malformed records found with -strict.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ChieNwosu/enterprise-incident-intelligence/internal/apachelog"
)

// Report is the top-level JSON document.
type Report struct {
	Input     string               `json:"input"`
	Summary   apachelog.Summary    `json:"summary"`
	Reference *apachelog.Agreement `json:"reference_agreement,omitempty"`
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("eii", flag.ContinueOnError)
	fs.SetOutput(stderr)
	in := fs.String("input", "", "path to Apache error log (required)")
	ref := fs.String("reference", "", "optional Loghub structured CSV for field agreement")
	out := fs.String("output", "", "JSON output path (default stdout)")
	top := fs.Int("top", 10, "number of top message patterns")
	maxLine := fs.Int("max-line-bytes", apachelog.DefaultMaxLineBytes, "maximum bytes per record; longer records are fatal")
	strict := fs.Bool("strict", false, "exit 3 if any malformed record is found")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *in == "" || *top < 0 || *maxLine <= 0 {
		fmt.Fprintln(stderr, "usage: eii -input FILE [-reference CSV] [-output JSON] [-top N] [-max-line-bytes N] [-strict]")
		return 2
	}

	f, err := os.Open(*in)
	if err != nil {
		fmt.Fprintf(stderr, "fatal: %v\n", err)
		return 1
	}
	defer f.Close()
	res, err := apachelog.Parse(f, *maxLine)
	if err != nil {
		fmt.Fprintf(stderr, "fatal: %v (read %d lines before failure; no summary written)\n", err, res.TotalLines)
		return 1
	}

	rep := Report{Input: *in, Summary: apachelog.Summarize(res, *top)}
	if *ref != "" {
		rf, err := os.Open(*ref)
		if err != nil {
			fmt.Fprintf(stderr, "fatal: %v\n", err)
			return 1
		}
		defer rf.Close()
		rows, err := apachelog.ReadReference(rf)
		if err != nil {
			fmt.Fprintf(stderr, "fatal: %v\n", err)
			return 1
		}
		a := apachelog.Compare(res, rows)
		rep.Reference = &a
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rep); err != nil {
		fmt.Fprintf(stderr, "fatal: %v\n", err)
		return 1
	}
	data := buf.Bytes()
	if *out == "" {
		_, err = stdout.Write(data)
	} else {
		err = os.WriteFile(*out, data, 0o644)
	}
	if err != nil {
		fmt.Fprintf(stderr, "fatal: write output: %v\n", err)
		return 1
	}
	fmt.Fprintf(stderr, "input=%d parsed=%d unparsed=%d\n", res.TotalLines, res.Parsed, res.Unparsed)
	if *strict && res.Unparsed > 0 {
		return 3
	}
	return 0
}
