package tests_utils

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

const (
	GrammarFile = "../grammar.txt"
	GrammarRoot = "Script"
)

var (
	grammarRule   = regexp.MustCompile(`^([A-Za-z]\w*)\s*=`)
	grammarQuoted = regexp.MustCompile(`'[^']*'`)
	grammarRef    = regexp.MustCompile(`\b[A-Z]\w*\b`)
)

type GrammarRule struct {
	Name string
	Body string
}

func (r GrammarRule) IsProduction() bool {
	return r.Name[0] >= 'A' && r.Name[0] <= 'Z'
}

func (r GrammarRule) Refs() []string {
	return grammarRef.FindAllString(grammarQuoted.ReplaceAllString(r.Body, ""), -1)
}

func ReadGrammar(t *testing.T, path string) []GrammarRule {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read grammar: %v", err)
	}

	var rules []GrammarRule
	var cur *GrammarRule
	for _, l := range strings.Split(string(data), "\n") {
		body := strings.TrimSpace(l)
		if cur == nil {
			m := grammarRule.FindStringSubmatch(l)
			if m == nil {
				continue
			}
			rules = append(rules, GrammarRule{Name: m[1]})
			cur = &rules[len(rules)-1]
			body = strings.TrimSpace(l[len(m[0]):])
		}
		cur.Body = strings.TrimSpace(cur.Body + " " + body)
		if strings.HasSuffix(body, ".") {
			cur = nil
		}
	}
	if cur != nil {
		t.Fatalf("rule %s is not terminated with .", cur.Name)
	}

	return rules
}

func RunGrammarClosed(t *testing.T, path string) {
	t.Helper()
	rules := ReadGrammar(t, path)
	defined := make(map[string]bool)
	used := map[string]bool{GrammarRoot: true}
	for _, r := range rules {
		if defined[r.Name] {
			t.Errorf("rule %s is defined twice", r.Name)
		}
		defined[r.Name] = true
		for _, ref := range r.Refs() {
			if ref != r.Name {
				used[ref] = true
			}
		}
	}
	for _, r := range rules {
		for _, ref := range r.Refs() {
			if !defined[ref] {
				t.Errorf("rule %s uses undefined %s", r.Name, ref)
			}
		}
		if r.IsProduction() && !used[r.Name] {
			t.Errorf("rule %s is never used", r.Name)
		}
	}
}

func RunGrammarCoverage(t *testing.T, grammarPath, srcPath string) {
	t.Helper()
	data, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	src := string(data)
	for _, r := range ReadGrammar(t, grammarPath) {
		if !r.IsProduction() {
			continue
		}
		t.Run(r.Name, func(t *testing.T) {
			if !strings.Contains(src, "// "+r.Name+" =") {
				t.Fatalf("no parse function documents %s = %s", r.Name, r.Body)
			}
		})
	}
}
