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

// srcFixtures copies the synthetic fixtures into a temp dir and returns paths.
func srcFixtures(t *testing.T) (dir, log, ref string) {
	t.Helper()
	dir = t.TempDir()
	log = filepath.Join(dir, "in.log")
	ref = filepath.Join(dir, "ref.csv")
	for src, dst := range map[string]string{"testdata/valid.log": log, "testdata/valid_ref.csv": ref} {
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return
}

func mustRead(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestOutputAliasesRejected(t *testing.T) {
	type mk func(t *testing.T, dir, log, ref string) string
	cases := []struct {
		name string
		mk   mk
	}{
		{"same as input", func(t *testing.T, dir, log, ref string) string { return log }},
		{"same as reference", func(t *testing.T, dir, log, ref string) string { return ref }},
		{"relative alias", func(t *testing.T, dir, log, ref string) string {
			t.Chdir(dir)
			return "./sub/../in.log"
		}},
		{"symlink to input", func(t *testing.T, dir, log, ref string) string {
			l := filepath.Join(dir, "link.json")
			if err := os.Symlink(log, l); err != nil {
				t.Skipf("symlinks unsupported: %v", err)
			}
			return l
		}},
		{"symlink to reference", func(t *testing.T, dir, log, ref string) string {
			l := filepath.Join(dir, "link.json")
			if err := os.Symlink(ref, l); err != nil {
				t.Skipf("symlinks unsupported: %v", err)
			}
			return l
		}},
		{"hard link to input", func(t *testing.T, dir, log, ref string) string {
			l := filepath.Join(dir, "hard.json")
			if err := os.Link(log, l); err != nil {
				t.Skipf("hard links unsupported: %v", err)
			}
			return l
		}},
		{"hard link to reference", func(t *testing.T, dir, log, ref string) string {
			l := filepath.Join(dir, "hard.json")
			if err := os.Link(ref, l); err != nil {
				t.Skipf("hard links unsupported: %v", err)
			}
			return l
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir, log, ref := srcFixtures(t)
			if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
				t.Fatal(err)
			}
			wantLog, wantRef := mustRead(t, log), mustRead(t, ref)
			out := tc.mk(t, dir, log, ref)
			code, stdout, errs := exec("-input", log, "-reference", ref, "-output", out)
			if code != 1 {
				t.Fatalf("code %d want 1: %s", code, errs)
			}
			if !strings.Contains(errs, "refusing to overwrite") || stdout != "" {
				t.Errorf("stderr %q stdout %q", errs, stdout)
			}
			if !bytes.Equal(mustRead(t, log), wantLog) || !bytes.Equal(mustRead(t, ref), wantRef) {
				t.Fatal("source file bytes changed")
			}
		})
	}
}

func TestOutputDistinctAccepted(t *testing.T) {
	_, log, ref := srcFixtures(t)
	wantLog, wantRef := mustRead(t, log), mustRead(t, ref)
	out := filepath.Join(filepath.Dir(log), "new.json")
	if code, _, errs := exec("-input", log, "-reference", ref, "-output", out); code != 0 {
		t.Fatalf("code %d: %s", code, errs)
	}
	if !json.Valid(mustRead(t, out)) {
		t.Error("output is not valid JSON")
	}
	// An existing distinct file is overwritten.
	if code, _, errs := exec("-input", log, "-reference", ref, "-output", out); code != 0 {
		t.Fatalf("rerun code %d: %s", code, errs)
	}
	if !bytes.Equal(mustRead(t, log), wantLog) || !bytes.Equal(mustRead(t, ref), wantRef) {
		t.Fatal("source file bytes changed")
	}
}

func TestStrictWritesReportThenExits3(t *testing.T) {
	out := filepath.Join(t.TempDir(), "o.json")
	code, _, _ := exec("-input", "testdata/malformed.log", "-strict", "-output", out)
	if code != 3 {
		t.Fatalf("code %d want 3", code)
	}
	if !json.Valid(mustRead(t, out)) {
		t.Error("strict run did not leave a valid report")
	}
}

func TestOutputIsDirectory(t *testing.T) {
	_, log, ref := srcFixtures(t)
	wantLog := mustRead(t, log)
	code, _, errs := exec("-input", log, "-reference", ref, "-output", t.TempDir())
	if code != 1 || !strings.Contains(errs, "fatal") {
		t.Fatalf("code %d: %s", code, errs)
	}
	if !bytes.Equal(mustRead(t, log), wantLog) {
		t.Fatal("source changed")
	}
}
