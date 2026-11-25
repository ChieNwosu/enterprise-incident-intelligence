package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func exec(args ...string) (int, string, string) {
	var o, e bytes.Buffer
	code := run(args, &o, &e)
	return code, o.String(), e.String()
}

func TestRunSuccess(t *testing.T) {
	out := filepath.Join(t.TempDir(), "o.json")
	code, _, errs := exec("-input", "testdata/valid.log", "-reference", "testdata/valid_ref.csv", "-output", out)
	if code != 0 {
		t.Fatalf("code %d: %s", code, errs)
	}
	if !strings.Contains(errs, "input=3 parsed=3 unparsed=0") {
		t.Errorf("stderr %q", errs)
	}
	b, _ := os.ReadFile(out)
	var rep Report
	if err := json.Unmarshal(b, &rep); err != nil {
		t.Fatal(err)
	}
	if rep.Summary.TotalLines != 3 || rep.Reference == nil || !rep.Reference.Aligned {
		t.Fatalf("%+v", rep)
	}
}

func TestRunStdout(t *testing.T) {
	code, out, _ := exec("-input", "testdata/valid.log")
	if code != 0 || !json.Valid([]byte(out)) {
		t.Fatalf("code %d out %q", code, out)
	}
}

func TestExitCodes(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"no args", nil, 2},
		{"bad flag", []string{"-nope"}, 2},
		{"missing file", []string{"-input", "testdata/does-not-exist.log"}, 1},
		{"oversized is fatal", []string{"-input", "testdata/oversized.log", "-max-line-bytes", "1024"}, 1},
		{"missing reference", []string{"-input", "testdata/valid.log", "-reference", "nope.csv"}, 1},
		{"malformed not fatal", []string{"-input", "testdata/malformed.log"}, 0},
		{"malformed strict", []string{"-input", "testdata/malformed.log", "-strict"}, 3},
		{"unwritable output", []string{"-input", "testdata/valid.log", "-output", "/nonexistent-dir/x.json"}, 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if code, _, _ := exec(tc.args...); code != tc.want {
				t.Fatalf("code %d want %d", code, tc.want)
			}
		})
	}
}

func TestFatalErrorDoesNotLeakContent(t *testing.T) {
	_, out, errs := exec("-input", "testdata/oversized.log", "-max-line-bytes", "1024")
	if out != "" || strings.Contains(errs, "AAAA") {
		t.Fatalf("leak or output: %q %q", out, errs)
	}
}
