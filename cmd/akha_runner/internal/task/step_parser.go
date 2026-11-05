package task

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Section struct {
	Name string
	Body string
	Line int
}

type Doc struct {
	Preamble string
	Sections []Section
	index    map[string]int
}

func (d *Doc) Get(name string) (string, bool) {
	i, ok := d.index[name]
	if !ok {
		return "", false
	}
	return d.Sections[i].Body, true
}

func (d *Doc) Has(name string) bool {
	_, ok := d.index[name]
	return ok
}

func (d *Doc) Names() []string {
	out := make([]string, len(d.Sections))
	for i, s := range d.Sections {
		out[i] = s.Name
	}
	return out
}

func ParseMDParseBytes(b []byte) (*Doc, error) { return parse(bytes.NewReader(b)) }

func ParseMDStepString(s string) (*Doc, error) { return parse(strings.NewReader(s)) }

func parse(r io.Reader) (*Doc, error) {
	doc := &Doc{index: map[string]int{}}

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 16*1024*1024)

	var (
		body      strings.Builder
		cur       = -1
		fenceChar byte
		fenceLen  int
		lineNo    int
	)

	flush := func() {
		text := strings.TrimSpace(body.String())
		if cur < 0 {
			doc.Preamble = text
		} else {
			doc.Sections[cur].Body = text
		}
		body.Reset()
	}

	for sc.Scan() {
		lineNo++
		line := strings.TrimSuffix(sc.Text(), "\r")

		if ch, n, ok := fence(line); ok {
			switch {
			case fenceChar == 0:
				fenceChar, fenceLen = ch, n
			case ch == fenceChar && n >= fenceLen && isClosingFence(line): // closing fence
				fenceChar, fenceLen = 0, 0
			}
		} else if fenceChar == 0 {
			if name, ok := sectionHeading(line); ok {
				if name == "" {
					return nil, fmt.Errorf("line %d: empty section name", lineNo)
				}
				if prev, dup := doc.index[name]; dup {
					return nil, fmt.Errorf("line %d: duplicate section %q (first at line %d)",
						lineNo, name, doc.Sections[prev].Line)
				}
				flush()
				doc.Sections = append(doc.Sections, Section{Name: name, Line: lineNo})
				cur = len(doc.Sections) - 1
				doc.index[name] = cur
				continue
			}
		}

		body.WriteString(line)
		body.WriteByte('\n')
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	flush()
	return doc, nil
}

func sectionHeading(line string) (string, bool) {
	s := trimIndent(line)
	if !strings.HasPrefix(s, "# ") && !strings.HasPrefix(s, "#\t") {
		return "", false
	}
	s = strings.TrimSpace(s[1:])
	s = strings.TrimSpace(strings.TrimRight(s, "#"))
	if !strings.HasPrefix(s, "@") {
		return "", false
	}
	return strings.TrimSpace(s[1:]), true
}

func fence(line string) (ch byte, n int, ok bool) {
	s := trimIndent(line)
	if len(s) < 3 || (s[0] != '`' && s[0] != '~') {
		return 0, 0, false
	}
	ch = s[0]
	for n < len(s) && s[n] == ch {
		n++
	}
	if n < 3 {
		return 0, 0, false
	}

	if ch == '`' && strings.ContainsRune(s[n:], '`') {
		return 0, 0, false
	}
	return ch, n, true
}

func isClosingFence(line string) bool {
	s := trimIndent(line)
	ch := s[0]
	return strings.TrimSpace(strings.TrimLeft(s, string(ch))) == ""
}

func trimIndent(line string) string {
	for i := 0; i < 3 && strings.HasPrefix(line, " "); i++ {
		line = line[1:]
	}
	return line
}
