// Command eii parses an Apache error log and writes a JSON summary.
//
// Exit codes: 0 success (malformed records are reported, not fatal),
// 1 fatal input/output error, 2 usage error, 3 malformed records found with -strict.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
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
	fset := flag.NewFlagSet("eii", flag.ContinueOnError)
	fset.SetOutput(stderr)
	in := fset.String("input", "", "path to Apache error log (required)")
	ref := fset.String("reference", "", "optional Loghub structured CSV for field agreement")
	out := fset.String("output", "", "JSON output path (default stdout)")
	top := fset.Int("top", 10, "number of top message patterns")
	maxLine := fset.Int("max-line-bytes", apachelog.DefaultMaxLineBytes, "maximum bytes per record; longer records are fatal")
	strict := fset.Bool("strict", false, "exit 3 if any malformed record is found")
	if err := fset.Parse(args); err != nil {
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
	srcs := []source{{"input", *in, nil}}
	if fi, err := f.Stat(); err == nil {
		srcs[0].info = fi
	} else {
		fmt.Fprintf(stderr, "fatal: stat input: %v\n", err)
		return 1
	}
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
		rfi, err := rf.Stat()
		if err != nil {
			fmt.Fprintf(stderr, "fatal: stat reference: %v\n", err)
			return 1
		}
		srcs = append(srcs, source{"reference", *ref, rfi})
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
		if err := checkOutput(*out, srcs); err != nil {
			fmt.Fprintf(stderr, "fatal: %v\n", err)
			return 1
		}
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

// source is an opened input file and its identity at open time.
type source struct {
	role string
	path string
	info os.FileInfo
}

// checkOutput rejects an output path that resolves (through relative paths,
// symbolic links or hard links) to one of the opened source files. A
// destination that does not exist yet is accepted. The check and the later
// write are not atomic: a concurrent actor that swaps the destination between
// them can defeat it.
func checkOutput(path string, srcs []source) error {
	// os.Stat follows symbolic links, so a link to a source is identified.
	fi, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("stat output %q: %w", path, err)
	}
	for _, s := range srcs {
		if os.SameFile(fi, s.info) {
			return fmt.Errorf("output %q is the same file as the %s file %q; refusing to overwrite a source", path, s.role, s.path)
		}
	}
	return nil
}
