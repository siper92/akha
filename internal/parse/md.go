package parse

import (
	"bufio"
	"errors"
	"strings"
)

const (
	ID_SECTION_KEY    = "id"
	TOOLS_SECTION_KEY = "tools"

	sectionPrefix = "# @"
)

var ErrMissingID = errors.New("task: missing '# @id: <name>' header")

type MDFlowStep struct {
	name     string
	sections []TextSection
}

func (s MDFlowStep) Name() string           { return s.name }
func (s MDFlowStep) Content() []TextSection { return s.sections }

func (s MDFlowStep) Section(id string) (TextSection, bool) {
	for _, sec := range s.sections {
		if sec.Title().ID() == id {
			return sec, true
		}
	}

	return nil, false
}

func Parse(doc string) (*MDFlowStep, error) {
	step := &MDFlowStep{}

	var (
		cur      *sectionBuilder
		sections []*sectionBuilder
	)

	sc := bufio.NewScanner(strings.NewReader(doc))
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if title, inline, ok := parseHeader(line); ok {
			cur = &sectionBuilder{title: title, header: line}
			if inline != "" {
				cur.body = append(cur.body, inline)
				cur.inline = true
			}
			sections = append(sections, cur)
			continue
		}
		if cur == nil {
			continue
		}
		cur.body = append(cur.body, line)
		cur.rawBody = append(cur.rawBody, line)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	for _, b := range sections {
		sec := b.build()
		if sec.title.ID() == ID_SECTION_KEY && step.name == "" {
			step.name = sec.content.String()
		}
		step.sections = append(step.sections, sec)
	}

	if step.name == "" {
		return nil, ErrMissingID
	}
	return step, nil
}

func parseHeader(line string) (title, inline string, ok bool) {
	if !strings.HasPrefix(line, sectionPrefix) {
		return "", "", false
	}
	rest := strings.TrimSpace(line[len(sectionPrefix):])
	if rest == "" {
		return "", "", false
	}
	if i := strings.Index(rest, ":"); i >= 0 {
		title = strings.TrimSpace(rest[:i])
		inline = strings.TrimSpace(rest[i+1:])
	} else {
		title = rest
	}
	if title == "" || strings.ContainsAny(title, " \t") {
		return "", "", false
	}
	return title, inline, true
}

type sectionBuilder struct {
	title   string
	header  string
	inline  bool
	body    []string
	rawBody []string
}

func (b *sectionBuilder) build() MDTextSection {
	raw := trimTrailingBlank(b.rawBody)
	content := trimBlank(b.body)

	rawText := b.header
	if len(raw) > 0 {
		rawText += "\n" + strings.Join(raw, "\n")
	}
	return MDTextSection{
		title:   MDTitle(b.title),
		content: Content(strings.Join(content, "\n")),
		raw:     rawText,
	}
}

func trimTrailingBlank(lines []string) []string {
	end := len(lines)
	for end > 0 && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return lines[:end]
}

func trimBlank(lines []string) []string {
	lines = trimTrailingBlank(lines)
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	return lines[start:]
}

func ParseTools(sec TextSection) map[string]string {
	tools := map[string]string{}
	for _, line := range strings.Split(sec.Content().String(), "\n") {
		name, desc, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if name = strings.TrimSpace(name); name != "" {
			tools[name] = strings.TrimSpace(desc)
		}
	}
	return tools
}
