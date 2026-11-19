package parse

const (
	IDEA_SECTION_KEY      = "idea"
	QUESTIONS_SECTION_KEY = "questions"
	ANSWERS_SECTION_KEY   = "answers"
	RATE_SECTION_KEY      = "rate"
)

type FlowStep interface {
	Name() string
	Content() []TextSection
}

type SectionContent interface {
	String() string
	Subsections() SubSectionContent
}

type SubSectionContent struct {
	name    string
	content []TextSection
}

func (s SubSectionContent) Name() string           { return s.name }
func (s SubSectionContent) Content() []TextSection { return s.content }

func (s SubSectionContent) Section(id string) (TextSection, bool) {
	for _, sec := range s.content {
		if sec.Title().ID() == id {
			return sec, true
		}
	}

	return nil, false
}

type SectionTitle interface {
	ID() string
	String() string
}

type TextSection interface {
	Title() SectionTitle
	Content() SectionContent
	String() string // title + content as in doc
}

var _ SectionTitle = (*MDTitle)(nil)

type MDTitle string

func (t MDTitle) ID() string {
	return string(t)
}

func (t MDTitle) String() string {
	return string(t)
}

var _ SectionContent = (*Content)(nil)

type Content struct {
	name string
	text string
}

func NewContent(name, text string) Content {
	return Content{name: name, text: text}
}

func (c Content) Subsections() SubSectionContent {
	return SubSectionContent{name: c.name, content: splitSubsections(c.text)}
}

func (c Content) String() string {
	return c.text
}

var _ TextSection = (*MDTextSection)(nil)

type MDTextSection struct {
	title   SectionTitle
	content SectionContent
	raw     string
}

func (M MDTextSection) Title() SectionTitle {
	return M.title
}

func (M MDTextSection) Content() SectionContent {
	return M.content
}

func (M MDTextSection) String() string {
	return M.raw
}
