package ignore

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestLiteralRegexOperators(t *testing.T) {
	for _, tc := range []struct {
		pattern string
		match   string
		miss    string
	}{
		{"c++", "c++", "cc"},
		{"build(1).log", "build(1).log", "build1.log"},
		{"foo|bar", "foo|bar", "bar"},
		{"^cache$", "^cache$", "cache"},
		{"part{2}", "part{2}", "partt"},
		{`\+cache`, "+cache", "cache"},
		{`build\(1\).log`, "build(1).log", "build1.log"},
		{"file[()+].txt", "file+.txt", "filex.txt"},
		{"file[^()+].txt", "filex.txt", "file+.txt"},
		{"Δ(λ)+.txt", "Δ(λ)+.txt", "Δλλ.txt"},
		{"log*/part{2}.txt", "logs/part{2}.txt", "logs/partt.txt"},
	} {
		t.Run(tc.pattern, func(t *testing.T) {
			base, err := ioutil.TempDir("", "ignore-literals-")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(base)
			path := filepath.Join(base, ".gitignore")
			if err := ioutil.WriteFile(path, []byte("# comment\n"+tc.pattern+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			fromFile, err := CompileIgnoreFile(path)
			if err != nil {
				t.Fatal(err)
			}
			fromFileAndLines, err := CompileIgnoreFileAndLines(path, "other.tmp")
			if err != nil {
				t.Fatal(err)
			}
			for _, matcher := range []*GitIgnore{CompileIgnoreLines("# comment", tc.pattern), fromFile, fromFileAndLines} {
				matched, how := matcher.MatchesPathHow(tc.match)
				if !matched || how == nil || how.Line != tc.pattern || how.LineNo != 2 {
					t.Errorf("MatchesPathHow(%q) = %t, %#v", tc.match, matched, how)
				}
				if matcher.MatchesPath(tc.miss) {
					t.Errorf("pattern %q matched %q", tc.pattern, tc.miss)
				}
			}
		})
	}
}

func TestLiteralRegexOperatorsInNegation(t *testing.T) {
	matcher := CompileIgnoreLines("*", "!build(1).log")
	if matcher.MatchesPath("build(1).log") {
		t.Error("literal exception remained ignored")
	}
	if !matcher.MatchesPath("build1.log") {
		t.Error("exception matched a different filename")
	}
}
